package main

import "time"
import "coord"
import "testing"
import "runtime"
import "strconv"
import "os"
import "fmt"

func port(tag string, host int) string {
	s := "/var/tmp/824-"
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

/* func TestSimple(t *testing.T) { */
/* 	runtime.GOMAXPROCS(4) */

/* 	const numTaskReplicas = 1 */

/* 	const nservers = 3 */
/* 	var coa []*coordinator.Coordinator = */ 
/*     make([]*coordinator.Coordinator, nservers) */
/* 	var kvh []string = make([]string, nservers) */
/* 	var sca []*coordinator.TestCoord = */
/*     make([]*coordinator.TestCoord, nservers) */
/* 	defer cleanup(coa) */

/* 	seed := int64(0) */

/* 	for i := 0; i < nservers; i++ { */
/* 		sca[i] = coordinator.MakeTestCoord() */
/* 	} */
/* 	for i := 0; i < nservers; i++ { */
/* 		kvh[i] = port("basic", i) */
/* 	} */
/* 	for i := 0; i < nservers; i++ { */
/* 		coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas, seed) */
/* 	} */

/*   numClient := 1; */
/* 	fmt.Printf("Test: Single Client\n") */

/*   clients := make([]*Client, numClient) */
/*   for i := 0; i < numClient; i++ { */
/*     options := &Options{ */
/*       kvh, */
/*       "/tmp/ts-clientsocket" + strconv.Itoa(i), */
/*       "./../../../libraries/client/python/testNode.py", */
/*     } */

/*     clients[i] = Init(options) */
/*   } */


/*   for _, c := range clients { */
/*     go c.Start() */
/*   } */

/*   // Time to finish the thing */
/*   time.Sleep(20 * time.Second) */

/* 	fmt.Printf("  ... Passed\n") */

/*   for _, c := range clients { */
/*     c.Kill() */
/*   } */

/* } */

/* func TestMultipleSimple(t *testing.T) { */
/* 	runtime.GOMAXPROCS(8) */

/* 	const numTaskReplicas = 1 */

/* 	const nservers = 3 */
/* 	var coa []*coordinator.Coordinator = */ 
/*     make([]*coordinator.Coordinator, nservers) */
/* 	var kvh []string = make([]string, nservers) */
/* 	var sca []*coordinator.TestCoord = */
/*     make([]*coordinator.TestCoord, nservers) */
/* 	defer cleanup(coa) */

/* 	seed := int64(0) */

/* 	for i := 0; i < nservers; i++ { */
/* 		sca[i] = coordinator.MakeTestCoord() */
/* 	} */
/* 	for i := 0; i < nservers; i++ { */
/* 		kvh[i] = port("basic", i) */
/* 	} */
/* 	for i := 0; i < nservers; i++ { */
/* 		coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas, seed) */
/* 	} */

/*   numClient := 7; */
/* 	fmt.Printf("Test: Multiple Clients\n") */

/*   clients := make([]*Client, numClient) */
/*   for i := 0; i < numClient; i++ { */
/*     options := &Options{ */
/*       kvh, */
/*       "/tmp/ts-clientsocket" + strconv.Itoa(i), */
/*       "./../../../libraries/client/python/testNode.py", */
/*     } */

/*     clients[i] = Init(options) */
/*   } */

/*   for _, c := range clients { */
/*     go c.Start() */
/*   } */

/*   // Time to finish the thing */
/*   time.Sleep(10 * time.Second) */

/*   for _, c := range clients { */
/*     c.Kill() */
/*   } */

/* 	fmt.Printf("  ... Passed\n") */
/* } */

/* func TestMultipleSimpleWithDelay(t *testing.T) { */
/* 	runtime.GOMAXPROCS(8) */

/* 	const numTaskReplicas = 1 */

/* 	const nservers = 3 */
/* 	var coa []*coordinator.Coordinator = */ 
/*     make([]*coordinator.Coordinator, nservers) */
/* 	var kvh []string = make([]string, nservers) */
/* 	var sca []*coordinator.TestCoord = */
/*     make([]*coordinator.TestCoord, nservers) */
/* 	defer cleanup(coa) */

/* 	seed := int64(0) */

/* 	for i := 0; i < nservers; i++ { */
/* 		sca[i] = coordinator.MakeTestCoord() */
/* 	} */
/* 	for i := 0; i < nservers; i++ { */
/* 		kvh[i] = port("basic", i) */
/* 	} */
/* 	for i := 0; i < nservers; i++ { */
/* 		coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas, seed) */
/* 	} */

/*   numClient := 7; */
/* 	fmt.Printf("Test: Multiple Clients with Join Delay\n") */

/*   clients := make([]*Client, numClient) */
/*   for i := 0; i < numClient; i++ { */
/*     options := &Options{ */
/*       kvh, */
/*       "/tmp/ts-clientsocket" + strconv.Itoa(i), */
/*       "./../../../libraries/client/python/testNode.py", */
/*     } */

/*     clients[i] = Init(options) */
/*   } */

/*   for _, c := range clients { */
/*     go c.Start() */
/*     time.Sleep(1 * time.Second) */
/*   } */

/*   // Time to finish the thing */
/*   time.Sleep(10 * time.Second) */

/*   for _, c := range clients { */
/*     c.Kill() */
/*   } */

/* 	fmt.Printf("  ... Passed\n") */
/* } */

func TestMultipleQuitThenJoin(t *testing.T) {
	runtime.GOMAXPROCS(8)

	const numTaskReplicas = 1

	const nservers = 3
	var coa []*coordinator.Coordinator = 
    make([]*coordinator.Coordinator, nservers)
	var kvh []string = make([]string, nservers)
	var sca []*coordinator.TestCoord =
    make([]*coordinator.TestCoord, nservers)
	defer cleanup(coa)

	seed := int64(0)

	for i := 0; i < nservers; i++ {
		sca[i] = coordinator.MakeTestCoord()
	}
	for i := 0; i < nservers; i++ {
		kvh[i] = port("basic", i)
	}
	for i := 0; i < nservers; i++ {
		coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas, seed)
	}

  numClient := 4;
	fmt.Printf("Test: Multiple Clients with Sync Join/Quit\n")

  clients := make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      "/tmp/ts-clientsocket" + strconv.Itoa(i),
      "./../../../libraries/client/python/testNode.py",
    }

    clients[i] = Init(options)
  }

  for _, c := range clients {
    go c.Start()
  }

  // Let them work for a bit, then kill them
  time.Sleep(4 * time.Second)

  for _, c := range clients {
    c.Kill()
  }

  // Make sure they're dead
  time.Sleep(1 * time.Second)

  // Start them again
  clients = make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      "/tmp/ts-clientsocket" + strconv.Itoa(i),
      "./../../../libraries/client/python/testNode.py",
    }

    clients[i] = Init(options)
  }

  for _, c := range clients {
    go c.Start()
  }

  // Let them work for a bit, then they should be done
  time.Sleep(5 * time.Second)

  for _, c := range clients {
    c.Kill()
  }

	fmt.Printf("  ... Passed\n")
}

func TestMultipleOOSQuitThenJoin(t *testing.T) {
	runtime.GOMAXPROCS(8)

	const numTaskReplicas = 1

	const nservers = 3
	var coa []*coordinator.Coordinator = 
    make([]*coordinator.Coordinator, nservers)
	var kvh []string = make([]string, nservers)
	var sca []*coordinator.TestCoord =
    make([]*coordinator.TestCoord, nservers)
	defer cleanup(coa)

	seed := int64(0)

	for i := 0; i < nservers; i++ {
		sca[i] = coordinator.MakeTestCoord()
	}
	for i := 0; i < nservers; i++ {
		kvh[i] = port("basic", i)
	}
	for i := 0; i < nservers; i++ {
		coa[i] = coordinator.StartServer(kvh, i, sca[i], numTaskReplicas, seed)
	}

  numClient := 4;
	fmt.Printf("Test: Multiple Clients with Out Of Sync Join/Quit\n")

  clients := make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      "/tmp/ts-clientsocket" + strconv.Itoa(i),
      "./../../../libraries/client/python/testNode.py",
    }

    clients[i] = Init(options)
  }

  for _, c := range clients {
    time.Sleep(time.Second)
    go c.Start()
  }

  for _, c := range clients {
    c.Kill()
    time.Sleep(time.Second)
  }

  // Start them again
  clients = make([]*Client, numClient)
  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      "/tmp/ts-clientsocket" + strconv.Itoa(i),
      "./../../../libraries/client/python/testNode.py",
    }

    clients[i] = Init(options)
  }

  for _, c := range clients {
    time.Sleep(time.Second)
    go c.Start()
  }

  // Let them work for a bit, then they should be done
  time.Sleep(5 * time.Second)

  for _, c := range clients {
    c.Kill()
  }

	fmt.Printf("  ... Passed\n")
}
