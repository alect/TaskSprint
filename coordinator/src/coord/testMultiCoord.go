package coordinator 

type SimpleMultiCoord struct { 
	currentTask TaskID 
	seed int64
	tasksFinished int
} 

func (sc *SimpleMultiCoord) Init(co *Coordinator, seed int64) { 
	// Start a task
	sc.seed = seed
	params := TaskParams { "test-task",  sc.seed, []string{"output"}, nil, nil, "Hello, world" } 
	sc.currentTask = co.StartTask(params)
} 

func (sc *SimpleMultiCoord) ClientJoined(co *Coordinator, CID ClientID, NumNodes int) { 
	// Do nothing for now 
} 

func (sc *SimpleMultiCoord) ClientDead(co *Coordinator, CID ClientID) { 
	// Do nothing for now 
} 

func (sc *SimpleMultiCoord) TaskDone(co *Coordinator, TID TaskID, DoneValues map[string]interface{}) { 
	// Handle tasks with prerequisites here 
	if TID == sc.currentTask { 
		sc.tasksFinished++ 
		if sc.tasksFinished > 10 { 
			co.Finish([]TaskID{sc.currentTask})
		} else { 
			params := TaskParams { "test-task", sc.seed, []string{"output"}, []TaskID{sc.currentTask}, []string{"internals"}, DoneValues["output"]}
			sc.currentTask = co.StartTask(params)
		} 
	}
} 

func MakeSimpleMultiCoord() *SimpleMultiCoord { 
	sc := new(SimpleMultiCoord)
	return sc
} 
