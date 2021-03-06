package coordinator 

import "log"
import "fmt"

type TestCoord struct {
	currentTask TaskID
	seed int64
	tasksFinished int
  results []int
  tasks map[TaskID]int
  numSubTasks int
  result int
}

func (sc *TestCoord) Init(co *Coordinator, seed int64) {
  sc.tasks = make(map[TaskID]int)
  sc.results = make([]int, 0)
  sc.numSubTasks = 400
	// Start a task
  for i := 0; i < sc.numSubTasks; i++ {
    s := i * 4
    args := []int{s, s + 1, s + 2, s + 3}
    sc.goPy(co, "sum", args, 1)
  }
}

func (sc *TestCoord) ClientJoined(co *Coordinator, CID ClientID, NumNodes int) {
	// Do nothing for now 
	fmt.Printf("Client joined: %v\n", CID)
}

func (sc *TestCoord) ClientDead(co *Coordinator, CID ClientID) {
	// Do nothing for now 
	fmt.Printf("Client marked dead: %v\n", CID)
}

func (sc *TestCoord) goPy(co *Coordinator, name string, args []int, taskType int) {
  params := TaskParams{}
  params.FuncName = name
  params.DoneKeys = []string{"result"}
  params.BaseObject = args
  tid := co.StartTask(params)
  sc.tasks[tid] = taskType
}

func (sc *TestCoord) TaskDone(co *Coordinator,
  TID TaskID, DoneValues map[string]interface{}) {
	// Handle tasks with prerequisites here 
  v, p := sc.tasks[TID]
  if !p {
    log.Fatal("TID was never created.")
  }

  result := int(DoneValues["result"].(float64))
  if v == 1 {
    sc.results = append(sc.results, result)

    if len(sc.results) == sc.numSubTasks {
      sc.goPy(co, "sum", sc.results, 2)
    }
	} else if v == 2 {
    /* fmt.Printf("Received: %d\n", result) */
    co.Finish([]TaskID{TID})
    sc.result = result
  }
}

// Returns 0 if not finished, > 0 when done
func (sc *TestCoord) Result() int {
  return sc.result
}

func MakeTestCoord() *TestCoord {
	sc := new(TestCoord)
	return sc
}

