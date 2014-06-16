package network

import (
	"bytes"
	"encoding/gob"
	"errors"
)

const (
	KNOWN_MAGIC = 0x987fc18e
)

type Frame struct {
	Peer    *Peer // Used for REP/REQ Pattern only
	Magic   uint32
	Type    string
	Payload []byte
}

func (f *Frame) GetBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(f)
	if err != nil {
		return nil
	} else {
		return buffer.Bytes()
	}
}

func NewFrame(frameType string, b []byte) *Frame {
	frame := new(Frame)
	frame.Magic = KNOWN_MAGIC
	frame.Type = frameType
	frame.Payload = make([]byte, len(b), len(b))
	copy(frame.Payload, b)
	return frame
}

func FrameFromBytes(b []byte) (*Frame, error) {
	var frame *Frame
	if len(b) < 12 {
		return frame, errors.New("Frame too short")
	}

	var buffer bytes.Buffer
	enc := gob.NewDecoder(&buffer)
	err := enc.Decode(&frame)
	if err != nil {
		return frame, err
	}

	return frame, nil
}
