package main

import "flag"
import "strings"
import "coord"
import "log"

func main() {
	servers := flag.String("servers", "", "comma-seperated list of servers")
	meIndex := flag.Int("me", 0, "The index in servers corresponding to this server")
	numTaskReplicas := flag.Int("n", 1, "The number of replicas we make for each task")
	seed := flag.Int64("seed", 0, "The seed assigned to this coordinator")
	developerCoord := flag.String("dc", "", "The requested developer coordinator")
	M := flag.Int("M", 1, "The M parameter for the map reduce coordinator")
	R := flag.Int("R", 1, "The R parameter for the map reduce coordinator")
	input := flag.String("input", "", "The input to the coordinator")

	flag.Parse()

	if *servers == "" || *meIndex == 0 || *developerCoord == "" {
		log.Fatal("usage: -servers ip:port[,ip:port] -me int -n int -seed int64 -dc coord")
	}

	serverNames := strings.Split(*servers, ",")

	// For now, just start a basic test coordinator
	var dc coordinator.DeveloperCoord
	if *developerCoord == "test" {
		dc = coordinator.MakeTestCoord()
	} else if *developerCoord == "mapreduce" {
		dc = coordinator.MakeMapReduceCoord(*M, *R, *input)
	} else {
		log.Fatal("dc parameter must be one of the following: test, mapreduce")
	}
	co := coordinator.MakeServer(serverNames, *meIndex, dc, *numTaskReplicas, *seed, "tcp")
	// Doesn't return until coordinator is dead. 
	co.StartTicks()
}