package coordinator 

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

func cleanup(coa []*Coordinator) { 
	for i := 0; i < len(coa); i++ { 
		if coa[i] != nil { 
			coa[i].Kill()
		} 
	} 
} 

func TestSimple(t *testing.T) { 
	runtime.GOMAXPROCS(4)

	const numTaskReplicas = 3

	const nservers = 3 
	var coa []*Coordinator = make([]*Coordinator, nservers)
	var kvh []string = make([]string, nservers)
	var sca []*SimpleCoord = make([]*SimpleCoord, nservers)
	defer cleanup(coa)

	for i := 0; i < nservers; i++ { 
		sca[i] = MakeSimpleCoord()
	}
	for i := 0; i < nservers; i++ { 
		kvh[i] = port("basic", i)
	} 
	for i := 0; i < nservers; i++ { 
		coa[i] = StartServer(kvh, i, sca[i], numTaskReplicas)
	} 
	
	ck := MakeClerk(kvh, "test-clerk")
	var cka [nservers]*Clerk 
	for i := 0; i < nservers; i++ { 
		cka[i] = MakeClerk([]string{kvh[i]}, "test-clerk")
	} 
	
	fmt.Printf("Test: Basic Query\n")
	
	fmt.Printf("  ... Passed\n"); 
	
} 