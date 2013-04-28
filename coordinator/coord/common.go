package coordinator 

type ClientID int64
type TaskID int64

type View struct { 
	// The struct representing the current view of the system 
	// used by the client to determine what it should be currently doing 
	
	// TODO: fill out the view
	TaskParams map[TaskID]TaskParams
	TaskAssignments map[TaskID][]ClientID
	
	ClientInfo map[ClientID]string // How to contact a particular client. 
} 

// To avoid concurrency issues when we're returning from a query, we clone the view.
func cloneView(oldView View) View { 
	newView := View{}
	newView.TaskParams = map[TaskID]TaskParams{}
	for k, v := range oldView.TaskParams { 
		newView.TaskParams[k] = v
	} 
	newView.TaskAssignments = map[TaskID][]ClientID{}
	for k, v := range oldView.TaskAssignments { 
		newView.TaskAssignments[k] = v
	} 
	newView.ClientInfo = map[ClientID]string 
	for k, v := range oldView.ClientInfo { 
		newView.ClientInfo[k] = v
	} 
	return newView
} 

type TaskParams struct { 
	TID TaskID 
	FuncName string 
	Seed int64
	DoneKeys []string 

	// The pre-requisite data needed to start this task 
	PreReqTasks []TaskID 
	PreReqKey string // The key for the data that should be requested of the pre-req tasks 
	BaseObject interface{} // Some base object for the task to start with 
} 


type QueryArgs struct { 
	CID ClientID
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