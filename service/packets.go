package service

import (
	"github.com/csanti/onet/network"
)

var WhisperType network.MessageTypeID
var AckType network.MessageTypeID
var IdaWhisperType network.MessageTypeID

func init() {
	WhisperType = network.RegisterMessage(&Whisper{})
	AckType = network.RegisterMessage(&Ack{})
	IdaWhisperType = network.RegisterMessage(&IdaWhisper{})
}

type Whisper struct {
	SourceId int // id of the node that generated the whisper
	Round int
	PeerId int // id of the node that gossiped the whisper
	Blob []byte // data
	Iteration int // increases everythime the mssage is regossiped
}

type IdaWhisper struct {
	SourceId int // id of the node that generated the whisper
	Round int
	PeerId int // id of the node that gossiped the whisper
	Blob []byte // data
	Iteration int // increases everythime the mssage is regossiped
	SegmentId int
}

type Ack struct {
	Timestamp string
	Round int
	PeerId int
	MaxIteration int
}