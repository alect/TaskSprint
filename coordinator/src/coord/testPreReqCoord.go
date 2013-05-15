package coordinator 

import "log"

type PreReqCoord struct {
  currentTask TaskID
  seed int64
  tasksFinished int
  results []int
  tasks map[TaskID]int
  numSubTasks int
  result int
}

func (sc *PreReqCoord) Init(co *Coordinator, seed int64) {
  sc.numSubTasks = 300
  sc.tasks = make(map[TaskID]int)
  sc.results = make([]int, 0)
  subtasks := make([]TaskID, sc.numSubTasks)
  keys := make([]string, sc.numSubTasks)

	// Start a task
  for i := 0; i < sc.numSubTasks; i++ {
    s := i * 4
    args := []int{s, s + 1, s + 2, s + 3}
    subtask := sc.goPy(co, "sum", args, 1, nil, nil)
    subtasks[i] = subtask
    keys[i] = "result"
  }

  sc.goPy(co, "sum", nil, 2, subtasks, keys)
}

func (sc *PreReqCoord) ClientJoined(co *Coordinator, CID ClientID, NumNodes int) {
	// Do nothing for now 
}

func (sc *PreReqCoord) ClientDead(co *Coordinator, CID ClientID) {
	// Do nothing for now 
	//fmt.Printf("Client marked dead: %v\n", CID)
}

func (sc *PreReqCoord) goPy(co *Coordinator, name string, args []int, taskType int, prereqTasks []TaskID, prereqKeys []string) TaskID {
  params := TaskParams{}
  params.FuncName = name
  if prereqTasks != nil {
    params.DoneKeys = []string{"result"}
  }
  params.PreReqTasks = prereqTasks
  params.PreReqKey = prereqKeys
  params.BaseObject = args
  task := co.StartTask(params)
  sc.tasks[task] = taskType
  return task
}

func (sc *PreReqCoord) TaskDone(co *Coordinator,
  TID TaskID, DoneValues map[string]interface{}) {
	// Handle tasks with prerequisites here 
  v, p := sc.tasks[TID]
  if !p {
    log.Fatal("TID was never created.")
  }

  if v == 2 {
    result := int(DoneValues["result"].(float64))
    co.Finish([]TaskID{TID})
    sc.result = result
  }
}

// Returns 0 if not finished, > 0 when done
func (sc *PreReqCoord) Result() int {
  return sc.result
}

func MakePreReqCoord() *PreReqCoord {
	sc := new(PreReqCoord)
	return sc
}

