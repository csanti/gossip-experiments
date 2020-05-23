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

	n := 20
	servers, roster, _ := test.GenTree(n, true)

	done := make(chan bool)
	var count int = 0
	cb := func() {
		count++
		if count >= n-1+1000 {
			done <- true
		}
	}

	gossipers := make([]*XGossip, n)
	for i := 0; i < n; i++ {
		c := &Config {
			Roster: roster,
			Index: i,
			N: n,
			CommunicationMode: 1,
			GossipPeers: 4,
			BlockSize: 10000000,
			RoundsToSimulate: 5,
			RoundTime: 15000,
			MinDelay: 100,
			MaxDelay: 600,
			UseSmart: true,
			MaxWeight: 100,
		}
		gossipers[i] = servers[i].Service(Name).(*XGossip)
		gossipers[i].SetConfig(c)
		gossipers[i].AttachCallback(cb)
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
