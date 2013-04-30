package coordinator 

type SimpleCoord struct { 
	co *Coordinator 
} 

func (sc *SimpleCoord) Init(co *Coordinator) { 
	// TODO: Start some tasks 
} 

func (sc *SimpleCoord) ClientJoined(co *Coordinator, CID ClientID) { 
	// Do nothing for now 
} 

func (sc *SimpleCoord) ClientDead(co *Coordinator, CID ClientID) { 
	// Do nothing for now 
} 

func (sc *SimpleCoord) TaskDone(co *Coordinator, TID TaskID, DoneValues map[string]interface{}) { 
	// TODO: Handle tasks with prerequisites here 
} 