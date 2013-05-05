package coordinator 

import "encoding/json"
import "fmt"
import "log"

type TestCoord struct {
	currentTask TaskID
	seed int64
	tasksFinished int
  results []int
  tasks map[TaskID]int
	finishedTasks map[TaskID]bool
  numSubTasks int
}

func (sc *TestCoord) Init(co *Coordinator, seed int64) {
  sc.tasks = make(map[TaskID]int)
	sc.finishedTasks = make(map[TaskID]bool)
  sc.results = make([]int, 0)
  sc.numSubTasks = 400
	// Start a task
  for i := 0; i < sc.numSubTasks; i++ {
    s := i * 4
    args := []int{s, s + 1, s + 2, s + 3}
    sc.goPy(co, "sum", args, 1)
  }
}

func (sc *TestCoord) ClientJoined(co *Coordinator, CID ClientID) {
	// Do nothing for now 
}

func (sc *TestCoord) ClientDead(co *Coordinator, CID ClientID) {
	// Do nothing for now 
	//fmt.Printf("Client marked dead: %v\n", CID)
}

func (sc *TestCoord) goPy(co *Coordinator, name string, args []int, taskType int) {
  params := TaskParams{}
  params.FuncName = name
  params.DoneKeys = []string{"result"}
  byteArray, err := json.Marshal(args)
  if err != nil {
    fmt.Println("JSON error.")
  }
  jsonString := string(byteArray)
  params.BaseObject = jsonString
  sc.tasks[co.StartTask(params)] = taskType
}

func (sc *TestCoord) TaskDone(co *Coordinator,
  TID TaskID, DoneValues map[string]interface{}) {
	// Handle tasks with prerequisites here 
  v, p := sc.tasks[TID]
  if !p {
    log.Fatal("TID was never created.")
  }

	sc.finishedTasks[TID] = true
  result := int(DoneValues["result"].(float64))
  if v == 1 {
    sc.results = append(sc.results, result)
		/*
		 if len(sc.results) < sc.numSubTasks && len(sc.results) > sc.numSubTasks - 5 {  
			for task, taskType := range sc.tasks { 
				_, taskFinished := sc.finishedTasks[task] 
				if !taskFinished && taskType == 1 { 
					fmt.Printf("Unfinished task: %v\n", task)
					view := co.GetCurrentView() 
					assts, assigned := view.TaskAssignments[task]
					if assigned { 
						fmt.Printf("Task %v assigned to %v\n", task, assts)
					} else { 
						fmt.Printf("Task %v not assigned\n", task)
					}
				} 
			}
		}*/ 

    if len(sc.results) == sc.numSubTasks {
      sc.goPy(co, "sum", sc.results, 2)
    }
	} else if v == 2 {
    fmt.Printf("Result from %d is %d\n", TID, result)
    co.Finish([]TaskID{TID})
  }
}

func MakeTestCoord() *TestCoord {
	sc := new(TestCoord)
	return sc
}

