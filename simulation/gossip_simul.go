package simulation

import (
	"time"

	"github.com/BurntSushi/toml"
	xgossip "github.com/csanti/gossip-experiments/service"
	"github.com/csanti/onet"
	"github.com/csanti/onet/log"
	//"github.com/csanti/onet/simul/monitor"
)

// Name is the name of the simulation
var Name = "xgossip"

func init() {
	onet.SimulationRegister(Name, NewSimulation)
}

// config being passed down to all nodes, each one takes the relevant
// information
type config struct {          // threshold of the threshold sharing scheme
	BlockSize int             // the size of the block in bytes         // blocktime in seconds
	GossipTime int
	GossipPeers int 		// number of neighbours that each node will gossip messages to
	CommunicationMode int  	// 0 for broadcast, 1 for gossip
	MaxRoundLoops int // maximum times a node can loop on a round before alerting
	RoundsToSimulate int
	RoundTime int
	MaxDelay int
	MinDelay int
	DefaultDelay int
	UseSmart bool
	MaxWeight int
	LoadTime int
	Timeout int
}

type Simulation struct {
	onet.SimulationBFTree
	config
}

func NewSimulation(config string) (onet.Simulation, error) {
	s := &Simulation{}
	_, err := toml.Decode(config, s)
	return s, err
}

func (s *Simulation) Setup(dir string, hosts []string) (*onet.SimulationConfig, error) {
	sim := new(onet.SimulationConfig)
	s.CreateRoster(sim, hosts, 2000)
	s.CreateTree(sim)
	return sim, nil
}

func (s *Simulation) DistributeConfig(config *onet.SimulationConfig) {
	n := len(config.Roster.List)
	for i, si := range config.Roster.List {
		c := &xgossip.Config{
			Roster: config.Roster,
			Index: i,
			N: n,
			CommunicationMode: s.CommunicationMode,
			GossipTime: s.GossipTime,
			GossipPeers: s.GossipPeers,
			BlockSize: s.BlockSize,
			MaxRoundLoops: s.MaxRoundLoops,
			RoundsToSimulate: s.RoundsToSimulate,
			RoundTime: s.RoundTime,
			MaxDelay: s.MaxDelay,
			MinDelay: s.MinDelay,
			DefaultDelay: s.DefaultDelay,
			UseSmart: s.UseSmart,
			MaxWeight: s.MaxWeight,
		}
		if i == 0 {
			config.GetService(xgossip.Name).(*xgossip.XGossip).SetConfig(c)
		} else {
			config.Server.Send(si, c)
		}

	}
}

func (s *Simulation) Run(config *onet.SimulationConfig) error {

	log.Lvl1("Distributing config to all nodes...")
	s.DistributeConfig(config)
	log.Lvl1("Sleeping for the config to dispatch correctly")
	time.Sleep(time.Duration(s.LoadTime) * time.Second)
	log.Lvl1("Starting xgossip simulation")
	xgossip := config.GetService(xgossip.Name).(*xgossip.XGossip)


	done := make(chan bool)

	newRoundCb := func(r int) {
		//roundDone++
		if r >= s.RoundsToSimulate-1 {
			done<-true
		}
		
		//log.Lvl1("Simulation round finished")
	}

	xgossip.AttachCallback(newRoundCb)

	xgossip.Start()

	select {
	case <-done:
	case <-time.After(time.Duration(s.Timeout) * time.Second):
		log.Lvl1("timeout")
	}

	log.Lvl1(" ---------------------------")
	log.Lvl1(" SIMULATION FINISHED ")
	log.Lvl1(" ---------------------------")
	return nil
}