package coordinator 

import "encoding/json"
import "fmt"

type TestCoord struct {
	currentTask TaskID
	seed int64
	tasksFinished int
}

func (sc *TestCoord) Init(co *Coordinator, seed int64) { 
	// Start a task
	sc.seed = seed
  params := TaskParams{}
  params.FuncName = "multiply"
  params.DoneKeys = []string{"result"}
  byteArray, err := json.Marshal([]int{1, 2})
  if err != nil {
    fmt.Println("JSON error.")
  }
  jsonString := string(byteArray)
  params.BaseObject = jsonString
	sc.currentTask = co.StartTask(params)
}

func (sc *TestCoord) ClientJoined(co *Coordinator, CID ClientID) { 
	// Do nothing for now 
}

func (sc *TestCoord) ClientDead(co *Coordinator, CID ClientID) { 
	// Do nothing for now 
}

func (sc *TestCoord) TaskDone(co *Coordinator,
  TID TaskID, DoneValues map[string]interface{}) {
	// Handle tasks with prerequisites here 
	if TID == sc.currentTask {
    byteArray := []byte(DoneValues["result"].(string))
    var result int
    json.Unmarshal(byteArray, result)
    fmt.Println("Result is %d\n", result)
    co.Finish([]TaskID{TID})
	}
}

func MakeTestCoord() *TestCoord {
	sc := new(TestCoord)
	return sc
}
