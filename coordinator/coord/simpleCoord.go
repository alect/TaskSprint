package coordinator 

type SimpleCoord struct { 
	co *Coordinator 
} 

func (sc *SimpleCoord) Init() { 
	// TODO: Start some tasks 
} 

func (sc *SimpleCoord) ClientJoined(CID ClientID) { 
	// Do nothing for now 
} 

func (sc *SimpleCoord) ClientDead(CID ClientID) { 
	// Do nothing for now 
} 

func (sc *SimpleCoord) TaskDone(TID TaskID, DoneValues map[string]interface{}) { 
	// TODO: Handle tasks with prerequisites here 
} 