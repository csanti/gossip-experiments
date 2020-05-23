package service

import (
	"github.com/csanti/onet"
	"github.com/csanti/onet/network"
)

var ConfigType network.MessageTypeID

func init() {
	ConfigType = network.RegisterMessage(&Config{})
}

// Config holds all the parameters for the consensus protocol
type Config struct {
	Roster *onet.Roster // participants
	Index  int          // index of the node receiving this config
	N      int          // length of participants
	GossipPeers int 		// number of neighbours that each node will gossip messages to
	CommunicationMode int  	// 0 for broadcast, 1 for gossip
	RoundsToSimulate int
	RoundTime int
	MaxDelay int
	MinDelay int
	UseSmart bool
	MaxWeight int
}
