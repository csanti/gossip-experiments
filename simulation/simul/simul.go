package main

import (
	// Service needs to be imported here to be instantiated.

	"github.com/csanti/onet/simul"
	_ "github.com/csanti/gossip-experiments/simulation"
)

func main() {
	simul.Start()
}