package service

import (
	"testing"
	"time"

	"github.com/csanti/onet"
	"github.com/csanti/onet/log"
	"go.dedis.ch/kyber"
	"go.dedis.ch/kyber/pairing"
)

type networkSuite struct {
	kyber.Group
	pairing.Suite
}

func newNetworkSuite() *networkSuite {
	return &networkSuite{
		Group: Suite.G2(),
		Suite: Suite,
	}
}

func TestXGossip(t *testing.T) {
	log.Lvl1("Starting test")

	suite := newNetworkSuite()
	test := onet.NewTCPTest(suite)
	defer test.CloseAll()

	n := 8
	servers, roster, _ := test.GenTree(n, true)

	done := make(chan bool)

	newRoundCb := func(r int) {
		//roundDone++
		if r >= 4 {
			done<-true
		}
		
		//log.Lvl1("Simulation round finished")
	}

	gossipers := make([]*XGossip, n)
	for i := 0; i < n; i++ {
		c := &Config {
			Roster: roster,
			Index: i,
			N: n,
			CommunicationMode: 1,
			GossipPeers: 3,
			RoundsToSimulate: 5,
			RoundTime: 25,
			MinDelay: 2000,
			MaxDelay: 4000,
			UseSmart: true,
			MaxWeight: 100,
			IdaGossipEnabled: true,
		}
		gossipers[i] = servers[i].Service(Name).(*XGossip)
		gossipers[i].SetConfig(c)
		gossipers[i].AttachCallback(newRoundCb)
	}

	time.Sleep(time.Duration(1)*time.Second)
	go gossipers[0].Start()
	select {
	case <-done:
	case <-time.After(1000 * time.Second):
		log.Lvl1("timeout")
	}
	
	log.Lvl1("Test finished.")
}
