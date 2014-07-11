package api

import (
	"emp/db"
	"emp/objects"
	"fmt"
	"quibit"
	"time"
)

func Start(config *ApiConfig) {
	var err error
	var frame quibit.Frame

	defer quit(config)

	// Start Database Services
	err = db.Initialize(config.Log, config.DbFile)
	defer db.Cleanup()
	if err != nil {
		config.Log <- fmt.Sprintf("Error initializing database: %s", err)
		config.Log <- "Quit"
		return
	}
	config.LocalVersion.Timestamp = time.Now().Round(time.Second)

	locVersion := objects.MakeFrame(objects.VERSION, objects.REQUEST, &config.LocalVersion)
	for str, _ := range config.NodeList.Nodes {
		locVersion.Peer = str
		config.SendQueue <- *locVersion
	}

	for {
		select {
		case frame = <-config.RecvQueue:
			if frame.Header.Command != objects.GETOBJ {
				config.Log <- fmt.Sprintf("Received %s frame...", CmdString(frame.Header.Command))
			} else {
				config.Log <- fmt.Sprintf("Received %s frame for %s...", CmdString(frame.Header.Command), frame.Payload)
			}
			switch frame.Header.Command {
			case objects.VERSION:
				version := new(objects.Version)
				err = version.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing version: %s", err)
				} else {
					go fVERSION(config, frame, version)
				}
			case objects.PEER:
				nodeList := new(objects.NodeList)
				err = nodeList.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing peer list: %s", err)
				} else {
					go fPEER(config, frame, nodeList)
				}
			case objects.OBJ:
				obj := new(objects.Obj)
				err = obj.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing obj list: %s", err)
				} else {
					go fOBJ(config, frame, obj)
				}
			case objects.GETOBJ:
				getObj := new(objects.Hash)
				if len(frame.Payload) == 0 {
					break
				}
				err = getObj.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing getobj hash: %s", err)
				} else {
					go fGETOBJ(config, frame, getObj)
				}
			case objects.PUBKEY_REQUEST:
				pubReq := new(objects.Hash)
				err = pubReq.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing pubkey request hash: %s", err)
				} else {
					go fPUBKEY_REQUEST(config, frame, pubReq)
				}
			case objects.PUBKEY:
				pub := new(objects.EncryptedPubkey)
				err = pub.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing pubkey: %s", err)
				} else {
					go fPUBKEY(config, frame, pub)
				}
			case objects.MSG:
				msg := new(objects.Message)
				err = msg.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing message: %s", err)
				} else {
					go fMSG(config, frame, msg)
				}
				fmt.Println("Finished select!")
			case objects.PURGE:
				purge := new(objects.Purge)
				err = purge.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing purge: %s", err)
				} else {
					go fPURGE(config, frame, purge)
				}
			default:
				config.Log <- fmt.Sprintf("Received invalid frame for command: %d", frame.Header.Command)
			}
		case <-config.Quit:
			fmt.Println()
			return
		}
	}

	// Should NEVER get here!
	panic("Must've been a cosmic ray!")
}

func quit(config *ApiConfig) {
	config.Log <- "Quit"
}
