package coordinator 

type SimpleCoord struct { 
	currentTask TaskID 
	seed int64
	tasksFinished int
} 

func (sc *SimpleCoord) Init(co *Coordinator, seed int64) { 
	// Start a task
	sc.seed = seed
	params := TaskParams { "test-task",  sc.seed, []string{"output"}, nil, nil, "Hello, world" } 
	sc.currentTask = co.StartTask(params)
} 

func (sc *SimpleCoord) ClientJoined(co *Coordinator, CID ClientID, NumNodes int) { 
	// Do nothing for now 
} 

func (sc *SimpleCoord) ClientDead(co *Coordinator, CID ClientID) { 
	// Do nothing for now 
} 

func (sc *SimpleCoord) TaskDone(co *Coordinator, TID TaskID, DoneValues map[string]interface{}) { 
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

func MakeSimpleCoord() *SimpleCoord { 
	sc := new(SimpleCoord)
	return sc
} 
