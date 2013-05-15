package main

import "time"
import "coord"
import "testing"
import "runtime"
import "strconv"
import "os"
import "fmt"
import "math/rand"

func port(tag string, host int, st string) string {
  if st != "unix" {
    port := 5000 + int(5000 * rand.Float64())
    return "localhost:" + strconv.Itoa(port)
  }
  s := "/tmp/824-"
  s += strconv.Itoa(os.Getuid()) + "/"
  os.Mkdir(s, 0777)
  s += "sm-"
  s += strconv.Itoa(os.Getpid()) + "-"
  s += tag + "-"
  s += strconv.Itoa(host)
  return s
}

func cleanup(coa []*coordinator.Coordinator) {
  for i := 0; i < len(coa); i++ {
    if coa[i] != nil {
      coa[i].Kill()
    }
  }
}

func Run(clients []*Client, nservers int, sca []*coordinator.TestCoord,
delay int, fail bool) {
  for _, c := range clients {
    go c.Start()
  }

  Poll(clients, nservers, sca, delay, 1279200, fail)

  for _, c := range clients {
    c.Kill()
  }

  // Make sure they're dead and sockets are freed
  time.Sleep(5 * time.Second)
}

func Poll(clients []*Client, nservers int, sca []*coordinator.TestCoord, 
delay int, expected int, fail bool) {
  endTime := time.Now().Add(time.Duration(delay) * time.Second)
  finished, timeout, result := 0, false, 0
  for finished < nservers && !timeout {
    finished = 0
    for _, c := range sca {
      result = c.Result()
      if result == 0 { continue }
      if result == expected { finished++
      } else { finished = 0; break; }
    }
    timeout = time.Now().After(endTime)
    time.Sleep(500 * time.Millisecond)
  }

  if fail {
    if !timeout && finished == nservers && result == expected {
      fmt.Printf("  ... Passed\n")
    } else if timeout {
      fmt.Printf("FAIL: {timeout, but got %d} \n", result)
    } else {
      fmt.Printf("FAIL: {expected %d, got %d}\n", expected, result)
    }
  }
}

func RunPreReq(clients []*Client, nservers int, sca []*coordinator.PreReqCoord,
delay int, fail bool) {
  for _, c := range clients {
    go c.Start()
  }

  PollPreReq(clients, nservers, sca, delay, 719400, fail)

  for _, c := range clients {
    c.Kill()
  }

  // Make sure they're dead and sockets are freed
  time.Sleep(5 * time.Second)
}

func PollPreReq(clients []*Client, nservers int,
sca []*coordinator.PreReqCoord, delay int, expected int, fail bool) {
  endTime := time.Now().Add(time.Duration(delay) * time.Second)
  finished, timeout, result := 0, false, 0
  for finished < nservers && !timeout {
    finished = 0
    for _, c := range sca {
      result = c.Result()
      if result == 0 { continue }
      if result == expected { finished++
      } else { finished = 0; break; }
    }
    timeout = time.Now().After(endTime)
    time.Sleep(500 * time.Millisecond)
  }

  if fail {
    if !timeout && finished == nservers && result == expected {
      fmt.Printf("  ... Passed\n")
    } else if timeout {
      fmt.Printf("FAIL: {timeout} \n", result)
    } else {
      fmt.Printf("FAIL: {expected %d, got %d}\n", expected, result)
    }
  }
}

func CreateCoords(nservers, numTaskReplicas int, seed int64,
st string) ([]*coordinator.Coordinator, []string, []*coordinator.TestCoord) {
  runtime.GOMAXPROCS(8)

  var coa []*coordinator.Coordinator = 
  make([]*coordinator.Coordinator, nservers)
  var kvh []string = make([]string, nservers)
  var sca []*coordinator.TestCoord =
  make([]*coordinator.TestCoord, nservers)

  for i := 0; i < nservers; i++ {
    sca[i] = coordinator.MakeTestCoord()
  }
  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i, st)
  }
  for i := 0; i < nservers; i++ {
    coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas,
      seed, st)
  }

  return coa, kvh, sca
}

func CreatePreReqCoords(nservers, numTaskReplicas int, seed int64,
st string) ([]*coordinator.Coordinator, []string, []*coordinator.PreReqCoord) {
  runtime.GOMAXPROCS(8)

  var coa []*coordinator.Coordinator = 
  make([]*coordinator.Coordinator, nservers)
  var kvh []string = make([]string, nservers)
  var sca []*coordinator.PreReqCoord =
  make([]*coordinator.PreReqCoord, nservers)

  for i := 0; i < nservers; i++ {
    sca[i] = coordinator.MakePreReqCoord()
  }
  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i, st)
  }
  for i := 0; i < nservers; i++ {
    coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas, seed,
    st)
  }

  return coa, kvh, sca
}

func CreatePythonCoords(nservers, numTaskReplicas int, seed int64,
st string, name string) ([]*coordinator.Coordinator,
[]string, []*coordinator.AllCoordinator) {
  runtime.GOMAXPROCS(8)

  var coa []*coordinator.Coordinator = 
  make([]*coordinator.Coordinator, nservers)
  var kvh []string = make([]string, nservers)
  var sca []*coordinator.AllCoordinator =
  make([]*coordinator.AllCoordinator, nservers)

  for i := 0; i < nservers; i++ {
    sca[i] = coordinator.MakeAllCoordinator(
      "./../../../libraries/python/coordinator/" + name + ".py");
  }
  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i, st)
  }
  for i := 0; i < nservers; i++ {
    coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas,
    seed, st)
  }

  return coa, kvh, sca
}

func RunPython(clients []*Client, nservers int, 
sca []*coordinator.AllCoordinator, delay int, fail bool) {
  RunPythonCustom(clients, nservers, sca, delay, fail, 7998000)
}

func RunPythonCustom(clients []*Client, nservers int, 
sca []*coordinator.AllCoordinator, delay int, fail bool, expected int) {
  for _, c := range clients {
    go c.Start()
  }

  PollPython(clients, nservers, sca, delay, expected, fail)

  for _, c := range clients {
    c.Kill()
  }

  // Make sure they're dead and sockets are freed
  time.Sleep(5 * time.Second)
}

func PollPython(clients []*Client, nservers int,
sca []*coordinator.AllCoordinator, delay int, expInt int, fail bool) {
  endTime := time.Now().Add(time.Duration(delay) * time.Second)
  finished, timeout, result, expected := 0, false, float64(0), float64(expInt)
  for finished < nservers && !timeout {
    finished = 0
    for _, c := range sca {
      result = c.Result()
      if result == 0 { continue }
      if result == expected || !fail { finished++
      } else { finished = 0; break; }
    }
    timeout = time.Now().After(endTime)
    time.Sleep(500 * time.Millisecond)
  }

  if fail {
    if !timeout && finished == nservers && result == expected {
      fmt.Printf("  ... Passed\n")
    } else if timeout {
      fmt.Printf("FAIL: {timeout} \n", result)
    } else {
      fmt.Printf("FAIL: {expected %d, got %d}\n", expected, result)
    }
  } else if result != 0 {
    fmt.Printf("  ... Passed (value: %f)\n", result)
  }
}

func CreateClients(numClient int, kvh []string, st string) []*Client {
  return CreateCustomClients(numClient, kvh, st, "testNode")
}

func CreateCustomClients(numClient int, kvh []string,
st string, program string) []*Client {
  clients := make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      port("clientsocket", i, st),
      st,
      "./../../../libraries/python/client/" + program + ".py",
    }

    clients[i] = Init(options)
  }

  return clients
}

func CreateMapReduceCoords(nservers, numTaskReplicas int, seed int64,
st string, name string) ([]*coordinator.Coordinator,
[]string, []*coordinator.AllCoordinator) {
  runtime.GOMAXPROCS(8)

  var coa []*coordinator.Coordinator = 
  make([]*coordinator.Coordinator, nservers)
  var kvh []string = make([]string, nservers)
  var sca []*coordinator.AllCoordinator =
  make([]*coordinator.AllCoordinator, nservers)

  for i := 0; i < nservers; i++ {
    sca[i] = coordinator.MakeAllCoordinator(
      "./../../../examples/mapreduce/mrfiles/" + name + ".py");
  }
  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i, st)
  }
  for i := 0; i < nservers; i++ {
    coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas,
    seed, st)
  }

  return coa, kvh, sca
}

func CreateMapReduceClients(numClient int, kvh []string,
st string, program string) []*Client {
  clients := make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      port("clientsocket", i, st),
      st,
      "./../../../examples/mapreduce/mrfiles/" + program + ".py",
    }
    clients[i] = Init(options)
  }

  return clients
}


func CreateMapReduceS3Coords(nservers, numTaskReplicas int, seed int64,
st string, name string) ([]*coordinator.Coordinator,
[]string, []*coordinator.AllCoordinator) {
  runtime.GOMAXPROCS(8)

  var coa []*coordinator.Coordinator = 
  make([]*coordinator.Coordinator, nservers)
  var kvh []string = make([]string, nservers)
  var sca []*coordinator.AllCoordinator =
  make([]*coordinator.AllCoordinator, nservers)

  for i := 0; i < nservers; i++ {
    sca[i] = coordinator.MakeAllCoordinator(
      "./../../../examples/mapreduce/mrs3/" + name + ".py");
  }
  for i := 0; i < nservers; i++ {
    kvh[i] = port("basic", i, st)
  }
  for i := 0; i < nservers; i++ {
    coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas,
    seed, st)
  }

  return coa, kvh, sca
}

func CreateMapReduceS3Clients(numClient int, kvh []string,
st string, program string) []*Client {
  clients := make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      port("clientsocket", i, st),
      st,
      "./../../../examples/mapreduce/mrs3/" + program + ".py",
    }
    clients[i] = Init(options)
  }

  return clients
}

func TestSimple(t *testing.T) {
  fmt.Printf("Test: Single Client\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 1;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "unix")
  clients := CreateClients(numClient, kvh, "unix")

  // Run the computation, timeout in 15 seconds
  Run(clients, nservers, sca, 15, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestSimpleTCP(t *testing.T) {
  fmt.Printf("Test: Single Client: TCP\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 1;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "tcp")
  clients := CreateClients(numClient, kvh, "tcp")

  // Run the computation, timeout in 15 seconds
  Run(clients, nservers, sca, 30, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMultipleSimple(t *testing.T) {
	fmt.Printf("Test: Multiple Clients\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 7;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "unix")
  clients := CreateClients(numClient, kvh, "unix")

  // Run the computation, timeout in 10 seconds
  Run(clients, nservers, sca, 10, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMultipleSimpleWithDelay(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with Join Delay\n")

  numTaskReplicas, nservers, numClient := 1, 3, 5;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "unix")
  clients := CreateClients(numClient, kvh, "unix")

  for _, c := range clients {
    go c.Start()
    time.Sleep(2 * time.Second)
  }

  Poll(clients, nservers, sca, 10, 1279200, true)

  for _, c := range clients {
    c.Kill()
  }

  cleanup(coa)
}

func TestMultipleQuitThenJoin(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with Sync Join/Quit\n")

  numTaskReplicas, nservers := 1, 3;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "unix")

  // Starting first batch of clients
  numClient := 4
  clients := CreateClients(numClient, kvh, "unix")
  Run(clients, nservers, sca, 5, false)

  // Start them again
  numClient = 4
  clients = CreateClients(numClient, kvh, "unix")
  Run(clients, nservers, sca, 10, true)

  // Cleanup coordinators
  cleanup(coa)
}

func TestOOSQuitThenJoin(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with Out Of Sync Join/Quit\n")
  numTaskReplicas, nservers, numClient := 1, 3, 4;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "unix")

  // First round
  clients := CreateClients(numClient, kvh, "unix")
  for _, c := range clients {
    time.Sleep(time.Second)
    go c.Start()
  }

  for _, c := range clients {
    c.Kill()
    time.Sleep(time.Second)
  }

  // Second round
  clients = CreateClients(numClient, kvh, "unix")
  for _, c := range clients {
    time.Sleep(time.Second)
    go c.Start()
  }

  // Let them work for a bit, then they should be done
  Poll(clients, nservers, sca, 5, 1279200, true)

  for _, c := range clients {
    c.Kill()
  }

  cleanup(coa)
}

func TestMultipleOOSQuitThenJoin(t *testing.T) {
  fmt.Printf("Test: Multiple Clients with Multiple OOS Join/Quit\n")
  numTaskReplicas, nservers := 1, 3;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "unix")

  startSleep := []time.Duration{250, 500, 300, 250, 0}
  killSleep := []time.Duration{500, 250, 800, 100, 0}
  numClients := []int{6, 4, 8, 2, 5}
  rounds := len(startSleep)
  for i := 0; i < rounds; i++ {
    clients := CreateClients(numClients[i], kvh, "unix")
    for _, c := range clients {
      time.Sleep(startSleep[i] * time.Millisecond)
      go c.Start()
    }

    if i == rounds - 1 {
      // Final round. Check for result
      Poll(clients, nservers, sca, 10, 1279200, true)
    }

    for _, c := range clients {
      c.Kill()
      time.Sleep(killSleep[i] * time.Millisecond)
    }

    // Let it rest
    time.Sleep(4 * time.Second)
  }

  cleanup(coa)
}

func TestMultipleOOSQuitThenJoinTCP(t *testing.T) {
  fmt.Printf("Test: Multiple Clients with Multiple OOS Join/Quit: TCP\n")
  numTaskReplicas, nservers := 1, 3;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "tcp")

  startSleep := []time.Duration{250, 500, 300, 250, 0}
  killSleep := []time.Duration{500, 250, 800, 100, 0}
  numClients := []int{2, 3, 4, 2, 3}
  rounds := len(startSleep)
  for i := 0; i < rounds; i++ {
    clients := CreateClients(numClients[i], kvh, "tcp")
    for _, c := range clients {
      time.Sleep(startSleep[i] * time.Millisecond)
      go c.Start()
    }

    if i == rounds - 1 {
      // Final round. Check for result
      Poll(clients, nservers, sca, 10, 1279200, true)
    }

    for _, c := range clients {
      c.Kill()
      time.Sleep(killSleep[i] * time.Millisecond)
    }

    // Let it rest
    time.Sleep(4 * time.Second)
  }

  cleanup(coa)
}

func TestSimpleLocalPreReq(t *testing.T) {
	fmt.Printf("Test: Single Client With Pre Reqs\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 1;
  coa, kvh, sca := CreatePreReqCoords(nservers, numTaskReplicas, 0, "unix")
  clients := CreateClients(numClient, kvh, "unix")

  // Run the computation, timeout in 15 seconds
  RunPreReq(clients, nservers, sca, 15, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestSimpleRemotePreReq(t *testing.T) {
	fmt.Printf("Test: Multiple Clients With Pre Reqs\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 3;
  coa, kvh, sca := CreatePreReqCoords(nservers, numTaskReplicas, 0, "unix")
  clients := CreateClients(numClient, kvh, "unix")

  // Run the computation, timeout in 15 seconds
  RunPreReq(clients, nservers, sca, 15, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestManySimpleRemotePreReq(t *testing.T) {
	fmt.Printf("Test: Many Clients With Pre Reqs\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 8;
  coa, kvh, sca := CreatePreReqCoords(nservers, numTaskReplicas, 0, "unix")
  clients := CreateClients(numClient, kvh, "unix")

  // Run the computation, timeout in 15 seconds
  RunPreReq(clients, nservers, sca, 15, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMultipleSimpleWithDelayAndPreReqs(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with Join Delay and Pre Reqs\n")

  numTaskReplicas, nservers, numClient := 1, 3, 5;
  coa, kvh, sca := CreatePreReqCoords(nservers, numTaskReplicas, 0, "unix")
  clients := CreateClients(numClient, kvh, "unix")

  for _, c := range clients {
    go c.Start()
    time.Sleep(2 * time.Second)
  }

  PollPreReq(clients, nservers, sca, 10, 719400, true)

  for _, c := range clients {
    c.Kill()
  }

  cleanup(coa)

  // Time to really clear open sockets/files
  time.Sleep(5 * time.Second)
}

func TestOOSQuitThenJoinPreReq(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with OOS Join/Quit and Pre Reqs\n")
  numTaskReplicas, nservers, numClient := 1, 3, 4;
  coa, kvh, sca := CreatePreReqCoords(nservers, numTaskReplicas, 0, "unix")

  // First round
  clients := CreateClients(numClient, kvh, "unix")
  for _, c := range clients {
    time.Sleep(time.Second)
    go c.Start()
  }

  for _, c := range clients {
    c.Kill()
    time.Sleep(time.Second)
  }

  // Second round
  clients = CreateClients(numClient, kvh, "unix")
  for _, c := range clients {
    time.Sleep(time.Second)
    go c.Start()
  }

  // Let them work for a bit, then they should be done
  PollPreReq(clients, nservers, sca, 5, 719400, true)

  for _, c := range clients {
    c.Kill()
  }

  cleanup(coa)

  // Time to really clear open sockets/files
  time.Sleep(5 * time.Second)
}

func TestMultipleOOSQuitThenJoinPreReq(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with Multiple OOS Join/Quit w/Pre Reqs\n")
  numTaskReplicas, nservers := 1, 3;
  coa, kvh, sca := CreatePreReqCoords(nservers, numTaskReplicas, 0, "unix")

  startSleep := []time.Duration{250, 500, 300, 250, 0}
  killSleep := []time.Duration{500, 250, 800, 100, 0}
  numClients := []int{6, 4, 8, 2, 5}
  rounds := len(startSleep)
  for i := 0; i < rounds; i++ {
    clients := CreateClients(numClients[i], kvh, "unix")
    for _, c := range clients {
      time.Sleep(startSleep[i] * time.Millisecond)
      go c.Start()
    }

    if i == rounds - 1 {
      // Final round. Check for result
      PollPreReq(clients, nservers, sca, 10, 719400, true)
    }

    for _, c := range clients {
      c.Kill()
      time.Sleep(killSleep[i] * time.Millisecond)
    }

    // Let it rest
    time.Sleep(5 * time.Second)
  }

  cleanup(coa)
}

func TestSimplePython(t *testing.T) {
	fmt.Printf("Test: Single Client With Python Coordinator\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 1;
  coa, kvh, sca := CreatePythonCoords(nservers, numTaskReplicas, 0, "unix",
    "testCoordinator")
  clients := CreateClients(numClient, kvh, "unix")

  // Run the computation, timeout in 15 seconds
  RunPython(clients, nservers, sca, 20, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMultiplePython(t *testing.T) {
	fmt.Printf("Test: Multiple Clients, Python Coordinator\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 5;
  coa, kvh, sca := CreatePythonCoords(nservers, numTaskReplicas, 0, "unix",
    "testCoordinator")
  clients := CreateClients(numClient, kvh, "unix")

  // Run the computation, timeout in 20 seconds
  RunPython(clients, nservers, sca, 15, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMultiplePythonTCP(t *testing.T) {
  fmt.Printf("Test: Multiple Clients, Python Coordinator: TCP\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 5;
  coa, kvh, sca := CreatePythonCoords(nservers, numTaskReplicas, 0, "tcp",
    "testCoordinator")
  clients := CreateClients(numClient, kvh, "tcp")

  // Run the computation, timeout in 15 seconds
  RunPython(clients, nservers, sca, 20, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMonteCarlo(t *testing.T) {
  fmt.Printf("Test: Monte Carlo\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 3;
  coa, kvh, sca := CreatePythonCoords(nservers, numTaskReplicas, 0, "unix",
    "mcCoord")
  clients := CreateCustomClients(numClient, kvh, "unix", "mcNode")

  // Run the computation, timeout in 10 seconds
  RunPython(clients, nservers, sca, 15, false)

  // Cleanup the coordinators
  cleanup(coa)
}


func TestRejoin(t *testing.T) {
	fmt.Printf("Test: Leave then rejoin\n")

	numTaskReplicas, nservers, numClient := 1, 3, 3
	coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "tcp")

	clients := CreateClients(numClient, kvh, "tcp")
	for _, c := range clients {
		fmt.Printf("CLient ID: %v\n", c.GetID())
	}
	Run(clients, nservers, sca, 2, false)
	
	fmt.Printf("Killing....\n")
	time.Sleep(12 * time.Second)
	fmt.Printf("Rejoining....\n")
	
	// Create clients with the same client IDs after they're dead
	clients_2 := CreateClients(numClient, kvh, "tcp")
	for i, c := range clients_2 {
		c.SetID(clients[i].GetID())
	}

	Run(clients_2, nservers, sca, 30, true)

	cleanup(coa)

}


func TestMapReducePython(t *testing.T) {
  fmt.Printf("Test: MapReduce Reverse Index\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 1
  coa, kvh, sca := CreateMapReduceCoords(nservers, numTaskReplicas, 0,
    "tcp", "mrCoord")
  clients := CreateMapReduceClients(numClient, kvh, "tcp", "mrReverseIndexNode")

  // Run the computation, timeout in 10 seconds
  RunPythonCustom(clients, nservers, sca, 60, false, 22222)

  // Cleanup the coordinators
  cleanup(coa)
}



// Uncomment this if you include your s3 credentials in private.py

/*
func TestMapReduceS3(t *testing.T) {
  fmt.Printf("Test: MapReduce Reverse Index on S3\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 1, 1
  coa, kvh, sca := CreateMapReduceS3Coords(nservers, numTaskReplicas, 0,
    "tcp", "mrCoord")
  clients := CreateMapReduceS3Clients(numClient, kvh, "tcp", "mrReverseIndexNode")

  // Run the computation, timeout in 10 seconds
  RunPythonCustom(clients, nservers, sca, 240, false, 22222)

  // Cleanup the coordinators
  cleanup(coa)
}
*/


/* func TestTaskReplication(t *testing.T) { */
/* 	fmt.Printf("Test: Task Replication with leaving clients\n") */
/* 	numTaskReplicas, nservers := 2, 3 */
/* 	coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0, "unix") */

/* 	clients_1 := CreateClients(2, kvh, "unix") */
/* 	clients_2 := CreateClients(2, kvh, "unix") */

/* 	for _, c := range clients_1 { */
/* 		go c.Start() */
/* 	} */
/* 	for _, c := range clients_2 { */
/* 		go c.Start() */
/* 	} */

/* 	Poll(clients_1, nservers, sca, 5, 1279200, false) */

/* 	for _, c := range clients_1 { */
/* 		c.Kill() */
/* 	} */

/* 	Poll(clients_2, nservers, sca, 20, 1279200, true) */

/* 	cleanup(coa) */
/* } */
