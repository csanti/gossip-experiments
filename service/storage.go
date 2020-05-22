package service

import (
	"time"

	"github.com/csanti/onet/network"
	"math/rand"
	"github.com/csanti/onet/log"
)

type SessionStorage struct {
	Peers map[int]*PeerInfo
	WeightedList []*PeerInfo
	InitialTimestamps map[int]time.Time
	MaxNodeDelay map[int]int
	MaxNodeIterations map[int]int
	AckCount map[int]int
	KnownPeers int
	AvgDelay int
	DelaySum int
	WeightSum int
	c *Config
	WeightDivider int
}

type PeerInfo struct {
	Index int
	Delay int //milliseconds
	*network.ServerIdentity
	Weight int
	NormWeight int
	Known bool
}

func NewSessionStorage(conf *Config) *SessionStorage {
	return &SessionStorage {
		Peers: make(map[int]*PeerInfo),
		MaxNodeDelay: make(map[int]int),
		MaxNodeIterations: make(map[int]int),
		InitialTimestamps: make(map[int]time.Time),
		AckCount: make(map[int]int),
		WeightSum: 10,
		c: conf,
		WeightDivider: 20,
	}
}

func (s *SessionStorage) InitPeerList() {
	for i, p := range s.c.Roster.List {
		s.Peers[i] = &PeerInfo {
			Index: i,
			ServerIdentity: p,
		}
		randomDelay := rand.Intn(s.c.MaxDelay - s.c.MinDelay) + s.c.MinDelay
		s.Peers[i].Delay = randomDelay
		s.DelaySum += randomDelay
	}
	s.AvgDelay = s.DelaySum / len(s.Peers)
	s.ComputeWeights()
}
/*
func (s *SessionStorage) PeerExists(id int) bool {
	if s.Peers[id] != nil {
		return true
	}
	return false
}


func (s *SessionStorage) UpdatePeer(id int, delay int) {
	if s.Peers[id] == nil {
		panic("Trying to update peer that does not exist")
	}
	s.Peers[id].Delay = delay
	s.Peers[id].Known = true
	if delay < s.MinDelay {
		s.MinDelay = delay
	} else if delay > s.MaxDelay {
		s.MaxDelay = delay
	}
	s.DelaySum += delay
	s.KnownPeers++
	s.AvgDelay = s.DelaySum / s.KnownPeers
}
*/

func (s *SessionStorage) ComputeWeights() {
	//s.WeightedList = make([]*PeerInfo, s.WeightSum)
	s.WeightedList = nil
	s.WeightSum = 0
	for _ , p := range s.Peers {

		p.Weight = int((1.0 / (float64(p.Delay) / float64(s.AvgDelay)))) 
		//p.Weight = int((1.0 / (float64(p.Delay) / float64(s.DelaySum)))/ float64(s.WeightDivider)) 
		//p.Weight = int(1.0 / (float64(p.Delay) / 100.0) * 10)
		//p.Weight = int(float64(p.Delay) / float64(s.DelaySum) * float64(s.WeightSum))	
		//log.Lvl1(p.Weight)
		//log.Lvl1(s.DelaySum)
		//log.Lvl1(s.WeightSum)
		s.WeightSum += p.Weight
		//log.Lvlf1("Source: %d - Node: %d - Delay: %d - Weight: %d",s.c.Index, p.Index, p.Delay, p.Weight)
	}

	for _ , p := range s.Peers {
			//s.WeightedList[index] = p
		p.NormWeight = int((float64(p.Weight)/float64(s.WeightSum)*float64(s.c.MaxWeight)))
		log.Lvlf1("Source: %d - Node: %d - Delay: %d - NormWeight: %d",s.c.Index, p.Index, p.Delay, p.NormWeight)
		for i:=0;i<p.NormWeight;i++ {

			s.WeightedList = append(s.WeightedList, p)
		}

	}
	//log.Lvl1(len(s.WeightedList))
	//s.WeightDivider++
	//log.Lvl1(s.Peers)
	//log.Lvl1(s.WeightedList)

}

func (s *SessionStorage) SmartSelectRandomPeer() *PeerInfo {
	rindex := rand.Intn(len(s.WeightedList))
	return s.WeightedList[rindex]
}	
func (s *SessionStorage) SelectRandomPeer() *PeerInfo {
	rindex := rand.Intn(len(s.Peers))
	return s.Peers[rindex]
}

func(s *SessionStorage) SetInitialTimestamp(r int, time time.Time) {
	s.InitialTimestamps[r] = time
}

func (s *SessionStorage) ProcessAck(a *Ack) int {
	ts, _ := time.Parse("2006-01-02 15:04:05.000000000 -0700 MS",a.Timestamp)
	delay := int(ts.Sub(s.InitialTimestamps[a.Round]))/1000000
	if delay > s.MaxNodeDelay[a.Round] {
		s.MaxNodeDelay[a.Round] = delay	
	}
	log.Lvl1(delay)
	if a.MaxIteration > s.MaxNodeIterations[a.Round] {
		s.MaxNodeIterations[a.Round] = a.MaxIteration
	}
	s.AckCount[a.Round]++
	return s.AckCount[a.Round]
}


