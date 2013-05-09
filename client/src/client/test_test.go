package main

import "time"
import "coord"
import "testing"
import "runtime"
import "strconv"
import "os"
import "fmt"

func port(tag string, host int) string {
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

  // Make sure they're dead
  time.Sleep(3 * time.Second)
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
      } else { finished = nservers; break; }
    }
    timeout = time.Now().After(endTime)
    time.Sleep(750 * time.Millisecond)
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

func CreateCoords(nservers, numTaskReplicas int,
seed int64) ([]*coordinator.Coordinator, []string, []*coordinator.TestCoord) {
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
    kvh[i] = port("basic", i)
  }
  for i := 0; i < nservers; i++ {
    coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas, seed)
  }

  return coa, kvh, sca
}

func CreateClients(numClient int, kvh []string) []*Client {
  clients := make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      port("clientsocket", i),
      "./../../../libraries/client/python/testNode.py",
    }

    clients[i] = Init(options)
  }

  return clients
}

func TestSimple(t *testing.T) {
	fmt.Printf("Test: Single Client\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 1;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0)
  clients := CreateClients(numClient, kvh)

  // Run the computation, timeout in 10 seconds
  Run(clients, nservers, sca, 30, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMultipleSimple(t *testing.T) {
	fmt.Printf("Test: Multiple Clients\n")

  // Set up coordinators and clients
  numTaskReplicas, nservers, numClient := 1, 3, 7;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0)
  clients := CreateClients(numClient, kvh)

  // Run the computation, timeout in 10 seconds
  Run(clients, nservers, sca, 10, true)

  // Cleanup the coordinators
  cleanup(coa)
}

func TestMultipleSimpleWithDelay(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with Join Delay\n")

  numTaskReplicas, nservers, numClient := 1, 3, 5;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0)
  clients := CreateClients(numClient, kvh)

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
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0)

  // Starting first batch of clients
  numClient := 4
  clients := CreateClients(numClient, kvh)
  Run(clients, nservers, sca, 4, false)

  // Start them again
  numClient = 4
  clients = CreateClients(numClient, kvh)
  Run(clients, nservers, sca, 8, true)

  // Cleanup coordinators
  cleanup(coa)
}

func TestOOSQuitThenJoin(t *testing.T) {
	fmt.Printf("Test: Multiple Clients with Out Of Sync Join/Quit\n")
  numTaskReplicas, nservers, numClient := 1, 3, 4;
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0)

  // First round
  clients := CreateClients(numClient, kvh)
  for _, c := range clients {
    time.Sleep(time.Second)
    go c.Start()
  }

  for _, c := range clients {
    c.Kill()
    time.Sleep(time.Second)
  }

  // Second round
  clients = CreateClients(numClient, kvh)
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
  coa, kvh, sca := CreateCoords(nservers, numTaskReplicas, 0)

  startSleep := []time.Duration{250, 500, 300, 250, 0}
  killSleep := []time.Duration{500, 250, 800, 100, 0}
  numClients := []int{6, 4, 8, 2, 5}
  rounds := len(startSleep)
  for i := 0; i < rounds; i++ {
    clients := CreateClients(numClients[i], kvh)
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
