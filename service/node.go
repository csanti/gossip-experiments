package service

import (
	"sync"
	"time"
	"math/rand"
	"github.com/csanti/onet"
	"github.com/csanti/onet/log"
	"github.com/csanti/onet/network"
)

type Node struct {
	*sync.Cond
	*onet.ServiceProcessor

	// config
	c *Config
	// current round number
	round int
	// done callback
	callback func(int) // callsback number of finalized blocks

	sessionStorage *SessionStorage
	isBlockProposer bool

	broadcast CommunicationFn
	gossipSubset CommunicationFn
}

func NewNodeProcess(c *onet.Context, conf *Config, b CommunicationFn, g CommunicationFn) *Node {
	// need to create chain first
	n := &Node {
		ServiceProcessor: onet.NewServiceProcessor(c),
		Cond: sync.NewCond(new(sync.Mutex)),
		c:     conf,
    	broadcast: b,
    	gossipSubset: g,
    	sessionStorage: NewSessionStorage(conf),
	}
	n.sessionStorage.InitPeerList()
	return n
}

func (n *Node) AttachCallback(fn func(int)) {
	// usually only attached to one of the nodes to notify a higher layer of the progress
	n.callback = fn
}

func (n *Node) Start() {
	log.Lvl1("Staring gossiping!")
	n.isBlockProposer = true
	// create first whisper
	packet := &Whisper{

	}
	// send bootstrap message to all nodes
	go n.gossip(packet)
}

func (n *Node) Process(e *network.Envelope) {
	n.Cond.L.Lock()
	defer n.Cond.L.Unlock()
	defer n.Cond.Broadcast()
	switch inner := e.Msg.(type) {
		case *Whisper:
			n.ReceivedWhisper(inner)
		default:
			log.Lvl2("Received unidentified message")
	}
}


func (n *Node) ReceivedWhisper(w *Whisper) {
	log.Lvl1("Processing whisper message...")

	// re gossip

	// start loop

	//go n.roundLoop(1)
}

func (n *Node) roundLoop(round int) {
	log.Lvlf3("Starting round %d loop",round)
	// 
	defer func() {
		log.Lvlf3("%d - Exiting round %d loop",n.c.Index,round)
		if n.callback != nil {
			n.callback(round)
		}
	}()
	n.Cond.L.Lock()
	defer n.Cond.L.Unlock()

	var times int = 0
	for {

		if times > n.c.MaxRoundLoops {
			log.Lvlf1("Node %d reached max round loops!!", n.c.Index)
			return
		}
		time.Sleep(time.Duration(n.c.GossipTime) * time.Millisecond)
		times++
	}
}

func (n *Node) gossip(msg interface{}) {
	for i := 0; i < n.c.GossipPeers; i++ {
		rp := n.sessionStorage.SelectRandomPeer()
		if !rp.Known {
			rp.Delay = rand.Intn(n.c.MaxDelay - n.c.MinDelay) + n.c.MinDelay
			n.sessionStorage.UpdatePeer(rp.Index, rp.Delay)
		}
		go func() {
			time.Sleep(time.Duration(rp.Delay) * time.Millisecond)
			if err := n.ServiceProcessor.SendRaw(rp.ServerIdentity, msg); err != nil {
				log.Lvl1("Error sending message")
			}		
		}()
	}
}