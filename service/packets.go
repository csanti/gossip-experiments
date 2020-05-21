package service

import (
	"github.com/csanti/onet/network"
)

var WhisperType network.MessageTypeID

func init() {
	WhisperType = network.RegisterMessage(&Whisper{})
}

type Whisper struct {
	SourceId int // id of the node that generated the whisper
	PeerId int // id of the node that gossiped the whisper
	Blob []byte // data
	Iteration int // increases everythime the mssage is regossiped
}