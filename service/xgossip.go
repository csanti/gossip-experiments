package service

import (
	"go.dedis.ch/kyber/pairing/bn256"
	"github.com/csanti/onet"
	"github.com/csanti/onet/log"
	"github.com/csanti/onet/network"
)

var Suite = bn256.NewSuite()
var G2 = Suite.G2()
var Name = "xgossip"

func init() {
	onet.RegisterNewService(Name, NewXGossipService)
}

// Dfinity service is either a beacon a notarizer or a block maker
type XGossip struct {
	*onet.ServiceProcessor
	context 	*onet.Context
	c       	*Config
	node    	*Node
}

func NewXGossipService(c *onet.Context) (onet.Service, error) {
	g := &XGossip{
		context:          c,
		ServiceProcessor: onet.NewServiceProcessor(c),
	}
	c.RegisterProcessor(g, WhisperType)
	c.RegisterProcessor(g, ConfigType)
	return g, nil
}

func (g *XGossip) SetConfig(c *Config) {
	g.c = c
	if g.c.CommunicationMode == 0 {
		g.node = NewNodeProcess(g.context, c, g.broadcast, g.gossip)
	} else if g.c.CommunicationMode == 1 {
		g.node = NewNodeProcess(g.context, c, g.broadcast, g.gossip)	
	} else {
		panic("Invalid communication mode")
	}	
}

func (g *XGossip) AttachCallback(fn func()) {
	// attach to something.. haha lol xd
	if g.node != nil {
		g.node.AttachCallback(fn)	
	} else {
		log.Lvl1("Could not attach callback, node is nil")
	}
}


func (g *XGossip) Start() {
	// send a bootstrap message
	if g.node != nil {
		g.node.Start()
	} else {
		panic("that should not happen")
	}
}

// Process
func (g *XGossip) Process(e *network.Envelope) {
	switch inner := e.Msg.(type) {
	case *Config:
		g.SetConfig(inner)
	case *Whisper:
		g.node.Process(e)
	default:
		log.Lvl1("Received unidentified message")
	}
}

// depreciated
func (n *XGossip) getRandomPeers(numPeers int) []*network.ServerIdentity {
	var results []*network.ServerIdentity
	for i := 0; i < numPeers; {
		posPeer := n.c.Roster.RandomServerIdentity()
		if n.ServerIdentity().Equal(posPeer) {
			// selected itself
			continue
		}
		results = append(results, posPeer)
		i++
	}
	return results
}

type CommunicationFn func(sis []*network.ServerIdentity, msg interface{})

func (g *XGossip) broadcast(sis []*network.ServerIdentity, msg interface{}) {
	for _, si := range sis {
		if g.ServerIdentity().Equal(si) {
			continue
		}
		log.Lvlf4("Broadcasting from: %s to: %s",g.ServerIdentity(), si)
		if err := g.ServiceProcessor.SendRaw(si, msg); err != nil {
			log.Lvl1("Error sending message")
			//panic(err)
		}
	}
}

func (g *XGossip) gossip(sis []*network.ServerIdentity, msg interface{}) {
	//targets := n.getRandomPeers(n.c.GossipPeers)
	targets := g.c.Roster.RandomSubset(g.ServerIdentity(), g.c.GossipPeers).List
	for k, target := range targets {
		if k == 0 {continue}
		log.Lvlf4("Gossiping from: %s to: %s",g.ServerIdentity(), target)
		if err := g.ServiceProcessor.SendRaw(target, msg); err != nil {
			log.Lvl1("Error sending message")
		}
	}
	
}
