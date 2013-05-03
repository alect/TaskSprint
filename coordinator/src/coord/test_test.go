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

	const numTaskReplicas = 1

	const nservers = 3 
	var coa []*Coordinator = make([]*Coordinator, nservers)
	var kvh []string = make([]string, nservers)
	var sca []*SimpleCoord = make([]*SimpleCoord, nservers)
	defer cleanup(coa)
	
	seed := int64(0)

	for i := 0; i < nservers; i++ { 
		sca[i] = MakeSimpleCoord()
	}
	for i := 0; i < nservers; i++ { 
		kvh[i] = port("basic", i)
	} 
	for i := 0; i < nservers; i++ { 
		coa[i] = StartServer(kvh, i, sca[i], numTaskReplicas, seed)
	} 
	
	ck := MakeClerk(kvh, "test-clerk", 4, 22)
	/*var cka [nservers]*Clerk 
	for i := 0; i < nservers; i++ { 
		cka[i] = MakeClerk([]string{kvh[i]}, "test-clerk", 4)
	}*/
	
	fmt.Printf("Test: Basic Query\n")
	
	view := ck.Query()
	// Check to make sure we've been assigned a task 
	if len(view.TaskAssignments) != 1 || len(view.TaskParams) != 1 || len(view.ClientInfo) != 1 { 
		t.Fatalf("Initial task assignment was wrong!\n")
	} 
	
	fmt.Printf("  ... Passed\n"); 

	fmt.Printf("Test: Finishing Task\n")
	
	// Pretend to finish the tasks 
	for tid := range view.TaskAssignments { 
		params := view.TaskParams[tid] 
		output := map[string]interface{}{}
		for _, key := range params.DoneKeys { 
			output[key] = "test"
		} 
		ck.Done(tid, output)
	} 

	view = ck.Query() 
	
	// Check to make sure a new task has been assigned 
	if len(view.TaskAssignments) != 2 || len(view.TaskParams) != 2 || len(view.ClientInfo) != 1 {	
		t.Fatalf("Completed tasks not functioning properly\n")
	} 

	params := view.TaskParams[TaskID(1)]
	if params.BaseObject != "test" { 
		t.Fatalf("Expected %v, received %v\n", "test", params.BaseObject)
	}

	fmt.Printf("  ... Passed\n")

	// Try to finish the tasks for this coordinator 
} 


func TestMulti(t *testing.T) { 
	runtime.GOMAXPROCS(4)

	const numTaskReplicas = 1

	const nservers = 3 
	var coa []*Coordinator = make([]*Coordinator, nservers)
	var kvh []string = make([]string, nservers)
	var sca []*SimpleMultiCoord = make([]*SimpleMultiCoord, nservers)
	defer cleanup(coa)
	
	seed := int64(0)

	for i := 0; i < nservers; i++ { 
		sca[i] = MakeSimpleMultiCoord()
	}
	for i := 0; i < nservers; i++ { 
		kvh[i] = port("basic", i)
	} 
	for i := 0; i < nservers; i++ { 
		coa[i] = StartServer(kvh, i, sca[i], numTaskReplicas, seed)
	} 
	
	//ck := MakeClerk(kvh, "test-clerk", 4, 22)
} 
