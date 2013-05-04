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

/* 	fmt.Printf("Test: Single Client\n") */

/*   options := &Options{ */
/*     kvh, */
/*     "/tmp/ts-clientsocket", */
/*     "./../../../libraries/client/python/testNode.py", */
/*   } */
/*   go Init(options) */

/*   time.Sleep(20 * time.Second) */

/* 	fmt.Printf("  ... Passed\n") */
/* } */

func TestMultipleSimple(t *testing.T) {
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

  numClient := 5;
	fmt.Printf("Test: Multiple Clients\n")

  for i := 0; i < numClient; i++ {
    options := &Options{
      kvh,
      "/tmp/ts-clientsocket" + strconv.Itoa(i),
      "./../../../libraries/client/python/testNode.py",
    }

    go Init(options)
  }

  time.Sleep(10 * time.Second)

	fmt.Printf("  ... Passed\n")
}
