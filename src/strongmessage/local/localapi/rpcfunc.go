package localapi

import (
	"errors"
	"fmt"
	"net/http"
	"strongmessage/encryption"
	"strongmessage/local/localdb"
	"strongmessage/objects"
)

var logChan chan string

func (s *StrongService) CreateAddress(r *http.Request, args *NilParam, reply *objects.AddressDetail) error {

	// Create Address

	priv, x, y := encryption.CreateKey(s.Config.Log)
	reply.Privkey = priv
	if x == nil {
		return errors.New("Key Pair Generation Error")
	}

	reply.Pubkey = encryption.MarshalPubkey(x, y)

	reply.IsRegistered = true

	reply.Address = encryption.GetAddress(s.Config.Log, x, y)

	if reply.Address == nil {
		return errors.New("Could not create address, function returned nil.")
	}

	reply.String = encryption.AddressToString(reply.Address)

	// Add Address to Database
	err := localdb.AddUpdateAddress(reply)
	if err != nil {
		s.Config.Log <- fmt.Sprintf("Error Adding Address: ", err)
		return err
	}

	// Send Pubkey to Network
	encPub := new(objects.EncryptedPubkey)

	encPub.AddrHash = objects.MakeHash(reply.Address)

	encPub.IV, encPub.Payload, err = encryption.SymmetricEncrypt(reply.Address, string(reply.Pubkey))
	if err != nil {
		s.Config.Log <- fmt.Sprintf("Error Encrypting Pubkey: ", err)
		return nil
	}

	// Record Pubkey for Network
	s.Config.RecvQueue <- *objects.MakeFrame(objects.PUBKEY, objects.BROADCAST, encPub)
	return nil
}

func (service *StrongService) GetAddress(r *http.Request, args *string, reply *objects.AddressDetail) error {
	var err error

	address := encryption.StringToAddress(*args)
	if len(address) != 25 {
		return errors.New(fmt.Sprintf("Invalid Address: %s", address))
	}

	addrHash := objects.MakeHash(address)

	detail, err := localdb.GetAddressDetail(addrHash)
	if err != nil {
		return err
	}

	// Check for pubkey
	if len(detail.Pubkey) == 0 {
		detail.Pubkey = checkPubkey(service.Config, objects.MakeHash(detail.Address))
	}

	*reply = *detail

	return nil
}

func (service *StrongService) AddUpdateAddress(r *http.Request, args *objects.AddressDetail, reply *NilParam) error {
	err := localdb.AddUpdateAddress(args)
	if err != nil {
		return err
	}

	checkPubkey(service.Config, objects.MakeHash(args.Address))

	return nil
}

func (service *StrongService) ListAddresses(r *http.Request, args *bool, reply *([]string)) error {
	strs := localdb.ListAddresses(*args)
	*reply = strs
	return nil
}
