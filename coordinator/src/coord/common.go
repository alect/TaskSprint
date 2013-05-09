package coordinator 

type ClientID int64
type TaskID int64
type NodeID int64

// The interface implemented by a developer coordinator 
type DeveloperCoord interface { 
	Init(co *Coordinator, seed int64)
	ClientJoined(co *Coordinator, CID ClientID) 
	ClientDead(co *Coordinator,  CID ClientID) 
	TaskDone(co *Coordinator, TID TaskID, DoneValues map[string]interface{})
} 

type View struct { 
	// The struct representing the current view of the system 
	// used by the client to determine what it should be currently doing 
	ViewNum int

	// Conclusion information 
	isFinished bool 
	outputTasks []TaskID

	// TODO: fill out the view
	
	// The old task info 
	TaskParams map[TaskID]TaskParams
	TaskAssignments map[TaskID][]ClientID
	FinishedTasks map[TaskID][]ClientID

	Tasks map[TaskID]TaskInfo

	ClientInfo map[ClientID]string // How to contact a particular client. 
} 

func (v *View) updateView() { 
	v.Tasks = map[TaskID]TaskInfo{}
	for tid, params := range v.TaskParams { 
		finishedClients := make([]ClientID, len(v.FinishedTasks[tid]))
		copy(finishedClients, v.FinishedTasks[tid])
		// Figure out what the pending clients are 
		pendingClients := make([]ClientID, len(v.TaskAssignments[tid]) - len(finishedClients))
		i := 0
		for _, cid := range v.TaskAssignments[tid] {
			// Check to see if this cid is already finished
			cidFinished := false
			for _, finishedCid := range finishedClients {
				if cid == finishedCid {
					cidFinished = true
					break
				}
			}
			if !cidFinished {
				pendingClients[i] = cid
				i++
			}
		}
		v.Tasks[tid] = TaskInfo{ params, pendingClients, finishedClients }
	}
}

// To avoid concurrency issues when we're returning from a query, we clone the view.
func cloneView(oldView View) View { 
	newView := View{}
	newView.ViewNum = oldView.ViewNum
	newView.TaskParams = map[TaskID]TaskParams{}
	for k, v := range oldView.TaskParams { 
		newView.TaskParams[k] = v
	} 
	newView.TaskAssignments = map[TaskID][]ClientID{}
	for k, v := range oldView.TaskAssignments {
		newView.TaskAssignments[k] = v
	}
	newView.FinishedTasks = map[TaskID][]ClientID{}
	for k, v := range oldView.FinishedTasks {
		newView.FinishedTasks[k] = v
	}
	newView.ClientInfo = map[ClientID]string{}
	for k, v := range oldView.ClientInfo {
		newView.ClientInfo[k] = v
	}
	// Build a new task info map out of the existing info 
	newView.Tasks = map[TaskID]TaskInfo{}
	for tid, params := range newView.TaskParams { 
		finishedClients := make([]ClientID, len(newView.FinishedTasks[tid]))
		copy(finishedClients, newView.FinishedTasks[tid])
		// Figure out what the pending clients are 
		pendingClients := make([]ClientID, len(newView.TaskAssignments[tid]) - len(finishedClients))
		i := 0
		for _, cid := range newView.TaskAssignments[tid] {
			// Check to see if this cid is already finished
			cidFinished := false
			for _, finishedCid := range finishedClients {
				if cid == finishedCid {
					cidFinished = true
					break
				}
			}
			if !cidFinished {
				pendingClients[i] = cid
				i++
			}
		}
		newView.Tasks[tid] = TaskInfo{ params, pendingClients, finishedClients }
	}
	return newView
}

type TaskParams struct { 
	FuncName string 
	Seed int64
	DoneKeys []string 

	// The pre-requisite data needed to start this task 
	PreReqTasks []TaskID 
	PreReqKey []string // The key for the data that should be requested of the pre-req tasks 
	BaseObject interface{} // Some base object for the task to start with 
} 

type TaskInfo struct { 
	Params TaskParams 
	PendingClients []ClientID
	FinishedClients []ClientID
} 



type QueryArgs struct { 
	CID ClientID
	Contact string // How to contact this client. Putting this here because I'm too lazy to figure out how to get it from the socket
	NumNodes int 
} 

type QueryReply struct { 
	View View
} 

type DoneArgs struct { 
	CID ClientID
	TID TaskID
	DoneValues map[string]interface{} // The values corresponding to the requested keys
} 

type DoneReply struct { 
} 
