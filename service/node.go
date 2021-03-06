package service

import (
	"sync"
	"time"
	"github.com/csanti/onet"
	"github.com/csanti/onet/log"
	"github.com/csanti/onet/network"
	"github.com/csanti/onet/simul/monitor"
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
	idaGossipEnabled bool

	broadcast CommunicationFn
	gossipSubset CommunicationFn
	simulatedRoundsCount int
	gossiped map[int]bool

	idaRounds map[int]*IdaRound

	finishedRound chan bool

	messagesSent map[int]int
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
    	gossiped: make(map[int]bool),
    	finishedRound: make(chan bool),
    	idaRounds: make(map[int]*IdaRound),
    	messagesSent: make(map[int]int),
	}
	n.sessionStorage.InitPeerList()
	return n
}

func (n *Node) AttachCallback(fn func(int)) {
	// usually only attached to one of the nodes to notify a higher layer of the progress
	n.callback = fn
}

func (n *Node) Start() {
	n.isBlockProposer = true
	go func() {
		for i:=0;i<n.c.RoundsToSimulate;i++ {
			log.Lvl1("Sending whisper ",i)

			// send bootstrap message to all nodes
			n.sessionStorage.SetInitialTimestamp(i, time.Now())

			if n.c.IdaGossipEnabled {

				go n.idaGossip(n.c.Index, i)
			} else {
				// create first whisper
				packet := &Whisper{
					SourceId: n.c.Index,
					Round: i,
					PeerId: n.c.Index,
					Iteration: 1,
				}
				n.round = i
				n.idaRounds[n.round] = NewIdaRound()
				go n.gossip(packet)
			}
			
			select {
				case <-n.finishedRound:
				case <-time.After(time.Duration(n.c.RoundTime) * time.Second):
					log.Lvl1("round timeout")
			}
			//n.finishedRound<-false
			log.Lvlf1("Round %d Finished - MaxDelay: %d - MaxIterations: %d - NodesReached: %d - SentMessages: %d ",i,n.sessionStorage.MaxNodeDelay[i],n.sessionStorage.MaxNodeIterations[i],n.sessionStorage.AckCount[i],n.messagesSent[i])
			if n.c.MonitoringEnabled {
				monitor.RecordSingleMeasure("maxDelay", float64(n.sessionStorage.MaxNodeDelay[i]))
				monitor.RecordSingleMeasure("maxIterations", float64(n.sessionStorage.MaxNodeIterations[i]))
				monitor.RecordSingleMeasure("reachedNodes", float64(n.sessionStorage.AckCount[i]))		
			}
			n.callback(i)
		}
	}()

}

func (n *Node) Process(e *network.Envelope) {
	n.Cond.L.Lock()
	defer n.Cond.L.Unlock()
	defer n.Cond.Broadcast()
	switch inner := e.Msg.(type) {
		case *Whisper:
			n.ReceivedWhisper(inner)
		case *IdaWhisper:
			n.ReceivedIdaWhisper(inner)
		case *Ack:
			n.ReceivedAck(inner)
		default:
			log.Lvl2("Received unidentified message")
	}
}

func (n *Node) ReceivedAck(a *Ack) {
	if n.sessionStorage.ProcessAck(a) == n.c.N  {
		n.finishedRound <- true
	}
}


func (n *Node) ReceivedWhisper(w *Whisper) {
	log.Lvl3("Processing whisper message...")
	if w.Round < n.round {
		// all nodes already finished, we dont need to keep gossiping
		return
	}
	if w.Round > n.round {
		n.round = w.Round
	}


	if !n.gossiped[w.Round] {
		received := time.Now()
		ack := &Ack {
			Timestamp: received.Format("2006-01-02 15:04:05.000000000 -0700 MS"),
			Round: w.Round,
			PeerId: n.c.Index,
			MaxIteration: w.Iteration,
		}

		w.PeerId = n.c.Index
		w.Iteration = w.Iteration + 1
		go n.gossip(w)
		n.gossiped[w.Round] = true

		if err := n.ServiceProcessor.SendRaw(n.c.Roster.Get(w.SourceId), ack); err != nil {
			log.Lvl1("Error sending ack")
		}	
	}
}

func (n *Node) ReceivedIdaWhisper(w *IdaWhisper) {
	//log.Lvl1("Processing ida whisper message... ", w.SegmentId, w.Round)
	if w.Round < n.round {
		// all nodes already finished, we dont need to keep gossiping
		return
	}
	if w.Round > n.round {
		n.round = w.Round
	}

	_, exists := n.idaRounds[w.Round]
	if !exists {
		n.idaRounds[w.Round] = NewIdaRound()
	}
	idaRound := n.idaRounds[w.Round]
	if !idaRound.GossipedSegments[w.SegmentId] {
		log.Lvl2("new segment ", n.round)
		idaRound.SegmentsRcv++
		idaRound.GossipedSegments[w.SegmentId] = true

		// re gossip segment
		w.PeerId = n.c.Index
		w.Iteration = w.Iteration + 1
		go n.gossip(w)

		if idaRound.SegmentsRcv >= n.c.GossipPeers - 2 { 
			// we have enough segments
			//log.Lvlf1("Node %d - Round %d - Received enough segments! ", n.c.Index, n.round)
			received := time.Now()
			ack := &Ack {
				Timestamp: received.Format("2006-01-02 15:04:05.000000000 -0700 MS"),
				Round: w.Round,
				PeerId: n.c.Index,
				MaxIteration: w.Iteration,
			}
			if err := n.ServiceProcessor.SendRaw(n.c.Roster.Get(w.SourceId), ack); err != nil {
				log.Lvl1("Error sending ack")
			}
		}
	}
}

/*
func (n *Node) roundLoop(round int) {
	log.Lvlf3("Starting round %d loop",round)
	// 
	defer func() {
		log.Lvlf3("%d - Exiting round %d loop",n.c.Index,round)
		if n.callback != nil {
			n.callback()
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
*/

func (n *Node) gossip(msg interface{}) {
	for i := 0; i < n.c.GossipPeers; i++ {
		var rp *PeerInfo
		if n.c.UseSmart {
			rp = n.sessionStorage.SmartSelectRandomPeer()
		} else {
			rp = n.sessionStorage.SelectRandomPeer()
		}		
		/*
		if !rp.Known {
			rp.Delay = rand.Intn(n.c.MaxDelay - n.c.MinDelay) + n.c.MinDelay
			n.sessionStorage.UpdatePeer(rp.Index, rp.Delay)
		} */
		//n.messagesSent[n.round]++
		go func() {
			//log.Lvl1(rp.Delay)
			time.Sleep(time.Duration(rp.Delay) * time.Millisecond)
			if err := n.ServiceProcessor.SendRaw(rp.ServerIdentity, msg); err != nil {
				log.Lvl1("Error sending message")
			}
		}()
	}
}

func (n *Node) idaGossip(index int, round int) {
	for i := 0; i < n.c.GossipPeers; i++ {
		var rp *PeerInfo
		if n.c.UseSmart {
			rp = n.sessionStorage.SmartSelectRandomPeer()
		} else {
			rp = n.sessionStorage.SelectRandomPeer()
		}

		idaWhisper := func(sId int, peer *PeerInfo) {
			packet := &IdaWhisper {
				SourceId: index,
				Round: round,
				PeerId: index,
				Iteration: 1,
				SegmentId: sId,
			}
			//log.Lvl1(rp.Delay)

			// reduce the dalay
			log.Lvlf1("%s - %d",peer.ServerIdentity, packet.SegmentId)
			time.Sleep(time.Duration(peer.Delay) * time.Millisecond)
			//n.messagesSent[round]++
			if err := n.ServiceProcessor.SendRaw(peer.ServerIdentity, packet); err != nil {
				log.Lvl1("Error sending message")
			}
		}
		go idaWhisper(i, rp)
	}
}