package coordinator 

/* import "fmt" */

/* type MapReduceCoord struct { */
/* 	M int // Number of map tasks to schedule */ 
/* 	R int // Number of reduce tasks to schedule */ 

/* 	input string // The input string for map-reduce */

/* 	mapTasks []TaskID */
/* 	reduceTasks []TaskID */
/* 	mergeOutputTask TaskID */
/* } */


/* func (mr *MapReduceCoord) Init(co *Coordinator, seed int64) { */
/* 	// Schedule all map tasks and all reduce tasks */ 
/* 	segmentLength := len(mr.input) / mr.M */
/* 	if segmentLength < 1 { */
/* 		segmentLength = 1 */
/* 	} */

/* 	// Start the map tasks */
/* 	for i := 0; i < mr.M; i++ { */
/* 		splitStart := i * segmentLength */
/* 		if splitStart >= len(mr.input) { */
/* 			splitStart = len(mr.input) */
/* 		} */
/* 		splitEnd := splitStart + segmentLength */
/* 		if splitEnd >= len(mr.input) { */
/* 			splitEnd = len(mr.input) */
/* 		} */

/* 		inputSplit := mr.input[splitStart:splitEnd] */
/* 		inputMap := map[string]interface{}{} */
/* 		inputMap["input"] = inputSplit */
/* 		inputMap["R"] = mr.R */
/* 		params := TaskParams { "map", 0, nil, nil, nil, inputMap } */
/* 		tid := co.StartTask(params) */
/* 		mr.mapTasks = append(mr.mapTasks, tid) */
/* 	} */
	
/* 	// Schedule the reduce tasks */
/* 	for i := 0; i < mr.R; i++ { */
/* 		params := TaskParams { "reduce", 0, nil, mr.mapTasks, []string{ fmt.SPrintf("reduce-%v", i) }, nil} */
/* 		tid := co.StartTask(params) */
/* 		mr.reduceTasks = append(mr.reduceTasks, tid) */
/* 	} */
	
/* 	// Finally, schedule the final merge task */
/* 	finalParams := TaskParams{"merge-output", 0, []string{"output"}, mr.reduceTasks, []string{"output"}, nil } */
/* 	mr.mergeOutputTask := co.StartTask(finalParams) */
/* } */

/* func (mr *MapReduceCoord) ClientJoined(co *Coordinator, CID ClientID) { */
/* 	// Do nothing */
/* } */

/* func (mr *MapReduceCoord) ClientDead(co *Coordinator, CID ClientID) { */
/* 	// Do nothing */
/* } */

/* func (mr *MapReduceCoord) TaskDone(co *Coordinator, TID TaskID, DoneValues map[string]interface{}) { */
/* 	if TID == mr.mergeOutputTask { */
/* 		// TODO: Do something with the output probably? */ 
		
/* 		co.Finish(nil) */
/* 	} */
/* } */

/* func MakeMapReduceCoord(M int, R int, input string) *MapReduceCoord { */ 
/* 	mr := new(MapReduceCoord) */ 
/* 	mr.M = M */
/* 	mr.R = R */ 
/* 	mr.input = input */
/* 	mr.mapTasks = make([]TaskID, 0) */
/* 	mr.reduceTasks = make([]TaskID, 0) */
/* 	mr.finishedTasks = map[TaskID]bool{} */
	

/* 	return mr */
/* } */ 
