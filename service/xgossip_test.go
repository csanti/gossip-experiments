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

	n := 10
	servers, roster, _ := test.GenTree(n, true)

	gossipers := make([]*XGossip, n)
	for i := 0; i < n; i++ {
		c := &Config {
			Roster: roster,
			Index: i,
			N: n,
			CommunicationMode: 1,
			GossipTime: 150,
			GossipPeers: 3,
			BlockSize: 10000000,
			MaxRoundLoops: 4,
			RoundsToSimulate: 10,
			DefaultDelay: 1000,
			MinDelay: 100,
			MaxDelay: 600,
		}
		gossipers[i] = servers[i].Service(Name).(*XGossip)
		gossipers[i].SetConfig(c)
	}
	
	done := make(chan bool)
	cb := func(r int) {
		if r > 10 {
			done <- true
		}
	}
	
	gossipers[0].AttachCallback(cb)
	time.Sleep(time.Duration(1)*time.Second)
	go gossipers[0].Start()
	<-done
	log.Lvl1("Test finished.")
}
