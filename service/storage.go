package service

import (
	"github.com/csanti/onet/network"
	"math/rand"
	"github.com/csanti/onet/log"
)

type SessionStorage struct {
	Peers map[int]*PeerInfo
	WeightedList []*PeerInfo
	KnownPeers int
	AvgDelay int
	MinDelay int
	MaxDelay int
	DelaySum int
	WeightSum int
	c *Config
}

type PeerInfo struct {
	Index int
	Delay int //milliseconds
	*network.ServerIdentity
	Weight int
	Known bool
}

func NewSessionStorage(conf *Config) *SessionStorage {
	return &SessionStorage {
		Peers: make(map[int]*PeerInfo),
		WeightSum: 10,
		c: conf,
	}
}

func (s *SessionStorage) InitPeerList() {
	for i, p := range s.c.Roster.List {
		s.Peers[i] = &PeerInfo {
			Index: i,
			ServerIdentity: p,
			Delay: s.c.DefaultDelay,
		}
	}
	s.DelaySum = len(s.Peers) * s.c.DefaultDelay
	s.ComputeWeights()
}

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

func (s *SessionStorage) ComputeWeights() {
	s.WeightedList = make([]*PeerInfo, s.WeightSum)
	var index int
	for _ , p := range s.Peers {
		p.Weight = int(float64(p.Delay) / float64(s.DelaySum) * float64(s.WeightSum))		
		for i := 0; i < p.Weight; i++ {
			s.WeightedList[index] = p
			index++
		}
	}
	log.Lvl1(s.Peers)
	log.Lvl1(s.WeightedList)

}

func (s *SessionStorage) SelectRandomPeer() *PeerInfo {
	rindex := rand.Intn(s.WeightSum)
	return s.WeightedList[rindex]
}	

