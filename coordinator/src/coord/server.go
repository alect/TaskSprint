package coordinator

import "net" 
import "fmt" 
import "net/rpc"
import "log" 
import "paxos" 
import "sync"
import "os" 
import "syscall" 
import "encoding/gob"
import "math/rand"
import "time"
import "container/list"

type Coordinator struct { 
	socktype string // The type of sockets used (unix for tests tcp for deployment)
	mu sync.Mutex
	l net.Listener
	me int 
	dead bool // For testing 
	unreliable bool // For testing 
	px *paxos.Paxos

	seed int64 // The random seed we share across all replicas

	dc DeveloperCoord // For performing callbacks. This is the developer defined object
	
	currentView View // The most recent view we're aware of
	initialized bool // Whether we've initialized yet. Calls init on first tick if not
	leaderID int // The id of the current leader. == me if we're the leader
	leaderNum int // The current leader number (not a leader id.)
	currentSeq int // The current sequence number (so we don't replay old log entries)

	nextTID TaskID // The next task ID we'll assign to a task

	// How long it's been since we last heard from a client in number of ticks
	lastQueries map[ClientID]int 

	// This state is unique to each replica 
	// Used to determine when replicas will attempt to elect a new leader
	lastLeaderElection time.Time

	// Used for task allocation 
	numTaskReplicas int // The number of replicas we make of each task. 
	unassignedTasks *list.List
	// The current assignment of active tasks 
	activeTasks map[ClientID]map[TaskID]bool
	// the number of nodes each client has available right now
	availableClients map[ClientID]int
	killedTasks map[TaskID]bool // Dead tasks that we shouldn't assign
	finishedTasks map[TaskID]bool // Tasks we've sent to the library indicating they've been finished

	// Output info for when the task is done 
	isFinished bool 
	outputTasks []TaskID

} 

func (co *Coordinator) GetCurrentView() View { 
	return co.currentView
} 

const (
	QUERY = 0
	DONE = 1
	TICK = 2
	LEADER_CHANGE = 3
	NIL = -1
)

const ELECTION_INTERVAL = 5 * time.Second

const DEAD_TICKS = 50 // Approximately 12 seconds

type PaxosReply struct { 
	View View
} 

type Op struct {
	// A Paxos log entry 
	Op int
	CID ClientID 
	Contact string 
	TID TaskID 
	DoneValues map[string]interface{}
	LeaderNum int
	LeaderID int
	NumNodes int
} 

// Have this function just so we don't compare the DoneValues when comparing log entries
func opsEqual(op1 Op, op2 Op) bool { 
	if op1.Op != op2.Op { 
		return false 
	} else if op1.Op == NIL { 
		return true 
	} else if op1.Op == TICK { 
		return op1.LeaderNum == op2.LeaderNum && op1.LeaderID == op2.LeaderID
	} else if op1.Op == LEADER_CHANGE { 
		return op1.LeaderNum == op2.LeaderNum && op1.LeaderID == op2.LeaderID
	} else if op1.Op == QUERY { 
		return op1.CID == op2.CID && op1.Contact == op2.Contact
	} else if op1.Op == DONE { 
		return op1.CID == op2.CID && op1.TID == op2.TID
	} 
	return false 
} 

var nilOp Op = Op{ Op: NIL }

func (co *Coordinator) WaitForPaxos(seq int, startFakeInst bool) Op { 
	to := 10 * time.Millisecond
	for !co.dead { 
		if seq < co.px.Min() { 
			return nilOp
		} 
		decided, v := co.px.Status(seq) 
		if decided { 
			return v.(Op)
		} else if startFakeInst { 
			co.px.Start(seq, nilOp)
			startFakeInst = false
		} 
		time.Sleep(to)
		if to < 10 * time.Second { 
			to *= 2
		} 
	} 
	return nilOp
}  

// Function that will apply an op assuming all previous ops in the log have been applied 
// Returns a view primarily for the sake of Query calls
func (co *Coordinator) ApplyPaxosOp (seq int, op Op) View { 
	// Only lock when we're applying ops 
	co.mu.Lock() 
	defer co.mu.Unlock()
	
	if (op.Op == NIL || seq <= co.currentSeq) {
		return co.currentView
	}
	oldViewNum := co.currentView.ViewNum
	co.currentSeq = seq
	// First, see if we need to initialize everything 
	if !co.initialized {
		co.dc.Init(co, co.seed)
		co.initialized = true
	}

	if op.Op == QUERY {
		// See if the client is new and assign tasks as appropriate
		_, exists := co.lastQueries[op.CID]
		if !exists {
			// new client event
			co.activeTasks[op.CID] = map[TaskID]bool{}
			co.availableClients[op.CID] = op.NumNodes
			// Update our view of how to contact the client 
			co.currentView.ClientInfo[op.CID] = op.Contact
			co.dc.ClientJoined(co, op.CID, op.NumNodes)
			// Increment our view
			co.currentView.ViewNum++
		}
		co.lastQueries[op.CID] = 0
		co.AllocateTasks()
		// Clone the current view for queries so concurrency doesn't affect it 
		//return cloneView(co.currentView)
	} else if op.Op == DONE {
		// Handle finished tasks 
		_, alreadyFinished := co.finishedTasks[op.TID]
		if !alreadyFinished {
			co.finishedTasks[op.TID] = true
			co.dc.TaskDone(co, op.TID, op.DoneValues)
		}
		// Check to see if this is even a living client
		_, alive := co.activeTasks[op.CID]
		// Check to see if this is an active task for this client 
		if alive {
			co.lastQueries[op.CID] = 0
			_, active := co.activeTasks[op.CID][op.TID] 
			if active {
				co.currentView.FinishedTasks[op.TID] = append(co.currentView.FinishedTasks[op.TID], op.CID)
				co.currentView.ViewNum++
				delete(co.activeTasks[op.CID], op.TID)
				co.availableClients[op.CID]++
			}
		}
		co.AllocateTasks()
	} else if op.Op == TICK {
		// Handle ticks to see if Some clients should be considered dead
		if op.LeaderNum < co.leaderNum {
			return co.currentView
		}
		for clientID, ticks := range co.lastQueries {
			if ticks+1 == DEAD_TICKS {
				// Dead client event
				co.ClientDead(clientID)
				co.dc.ClientDead(co, clientID)
				co.currentView.ViewNum++

				// If we don't have any more clients and we're done, shut down this replica 
				if co.isFinished && len(co.availableClients) == 0 {
					co.Kill()
				}
				// TEST: Make sure the client is actually dead 
				_, stillAlive := co.lastQueries[clientID]
				if stillAlive {
					log.Fatal("Failed to properly kill client")
				}

			} else {
				co.lastQueries[clientID]++
			}
		}

		co.AllocateTasks()
	} else if op.Op == LEADER_CHANGE {
		// Handle a leader change 
		// Only change leaders if the new leader is actually new
		if op.LeaderNum > co.leaderNum { 
			co.leaderNum = op.LeaderNum
			co.leaderID = op.LeaderID
			co.lastLeaderElection = time.Now()
		}
	}
	if co.currentView.ViewNum > oldViewNum {
		(&co.currentView).updateView()
	}
	return co.currentView
}

// Function for driving the paxos log forward if we've lagged a bit 
func (co *Coordinator) UpdatePaxos() {
	max := co.px.Max()
	for seq := co.px.Min(); seq <= max && !co.dead; seq++ {
		oldOp := co.WaitForPaxos(seq, true)
		co.ApplyPaxosOp(seq, oldOp)
	}
	co.px.Done(max)
}

// Function that attempts to insert an op into the paxos log and 
// applies each op as it's discovered 
func (co *Coordinator) PerformPaxos(op Op) View {
	// First, catch up if necessary
	max := co.px.Max()
	for seq := co.px.Min(); seq <= max && !co.dead; seq++ {
		oldOp := co.WaitForPaxos(seq, true)
		co.ApplyPaxosOp(seq, oldOp)
	}
	co.px.Done(max)

	// Perform Paxos until we succeed 
	for seq := max+1; !co.dead; seq++ {
		co.px.Start(seq, op)
		v := co.WaitForPaxos(seq, false)
		// Apply the op
		view := co.ApplyPaxosOp(seq, v)
		co.px.Done(seq)
		if opsEqual(op, v) {
			return view
		}
	}
	return View{}
}

// Used to change leaders and to update the view of who's dead if we're the leader 
func (co *Coordinator) tick() { 

	co.UpdatePaxos()

	// First, see if it's time to elect a new leader 
	co.mu.Lock() 
	shouldElectNewLeader := time.Since(co.lastLeaderElection) >= ELECTION_INTERVAL
	nextLeaderNum := co.leaderNum+1
	co.mu.Unlock() 

	if shouldElectNewLeader { 
		leaderOp := Op { Op: LEADER_CHANGE, LeaderNum: nextLeaderNum, LeaderID: co.me }
		co.PerformPaxos(leaderOp)
	} 

	// See if we need to insert a tick 
	co.mu.Lock() 
	shouldInsertTick := co.leaderID == co.me
	leaderID := co.leaderID
	leaderNum := co.leaderNum
	co.mu.Unlock()

	if shouldInsertTick { 
		tickOp := Op { Op: TICK, LeaderNum: leaderNum, LeaderID: leaderID }
		co.PerformPaxos(tickOp)
	} 
} 

// Called automatically to allocate tasks to clients as clients become available
func (co *Coordinator) AllocateTasks() {
	// Go through each of our available clients and attempt to allocate a task
	for cid, N := range co.availableClients {
		for i := 0; i < N; i++ {
			found := false
			var tid TaskID
			for e := co.unassignedTasks.Front(); e != nil; e = e.Next() {
				tid = e.Value.(TaskID)
				// Check to see if it's been killed
				_, killed := co.killedTasks[tid]
				if killed {
					co.unassignedTasks.Remove(e)
					continue
				}
				// Check to see if it's already assigned 
				alreadyAssigned := false
				for _, assCid := range co.currentView.TaskAssignments[tid] {
					if assCid == cid {
						alreadyAssigned = true
					}
				}
				if !alreadyAssigned {
					found = true
					co.unassignedTasks.Remove(e)
					break
				}
			}
			
			if found {
				co.currentView.ViewNum++
				//fmt.Printf("Assigning Task %v to Client %v\n", tid, cid)
				co.currentView.TaskAssignments[tid] = append(co.currentView.TaskAssignments[tid], cid)
				co.activeTasks[cid][tid] = true
				co.availableClients[cid]--
			} else {
				break
			}
		}
	}
}

// When a client dies, have to remove it from the view and possibly 
// re-assign its tasks 
func (co *Coordinator) ClientDead(CID ClientID) { 
	// First, this client is not available 
	delete(co.availableClients, CID) 
	delete(co.activeTasks, CID)
	delete(co.lastQueries, CID)
	delete(co.currentView.ClientInfo, CID)

	// Now, find tasks this client is responsible for and remove the client 
	for tid, clients := range co.currentView.TaskAssignments {
		foundClient := false 
		newAsst := make([]ClientID, 0)
		for _, client := range clients { 
			if client == CID { 
				foundClient = true 
			} else { 
				newAsst = append(newAsst, client)
			} 
		} 
		if foundClient { 
			co.currentView.TaskAssignments[tid] = newAsst
			// reassign this task?
			co.unassignedTasks.PushFront(tid)
		}
	}
	for tid, clients := range co.currentView.FinishedTasks {
		foundClient := false 
		newFinished := make([]ClientID, 0)
		for _, client := range clients {
			if client == CID {
				foundClient = true
			} else {
				newFinished = append(newFinished, client)
			}
		}
		if foundClient {
			co.currentView.FinishedTasks[tid] = newFinished
		}
	}
}


// RPC functions 

// When a client wants the latest view
func (co *Coordinator) Query(args *QueryArgs, reply *QueryReply) error { 
	op := Op { Op: QUERY, CID: args.CID, Contact: args.Contact, NumNodes: args.NumNodes }
	result := co.PerformPaxos(op)
	reply.View = result
	return nil 
} 

// When a client has finished a task 
func (co *Coordinator) TaskDone(args *DoneArgs, reply *DoneReply) error { 
	op := Op { Op: DONE, CID: args.CID, TID: args.TID, DoneValues: args.DoneValues }
	co.PerformPaxos(op)
	return nil
} 


// Functions called by the developer coordinator to manage tasks 
func (co *Coordinator) StartTask(params TaskParams) TaskID { 
	// Function that starts a task 
	tid := co.nextTID
	co.nextTID++
	co.currentView.TaskParams[tid] = params
	co.currentView.TaskAssignments[tid] = make([]ClientID, 0)
	co.currentView.FinishedTasks[tid] = make([]ClientID, 0)
	co.currentView.ViewNum++
	// Add this task to a list of unassigned tasks 
	// It will be assigned to clients as appropriate
	for i := 0; i < co.numTaskReplicas; i++ { 
		co.unassignedTasks.PushBack(tid)
	} 
	return tid
} 

func (co *Coordinator) KillTask(tid TaskID) { 
	// Kills a task by removing its assignment from the view
	// Can be used to allow clients to discard data that's no longer needed
	delete(co.currentView.TaskParams, tid)
	delete(co.currentView.FinishedTasks, tid)
	// Release clients for whom the task is currently active
	for _, cid := range co.currentView.TaskAssignments[tid] {
		_, active := co.activeTasks[cid][tid]
		if active {
			co.availableClients[cid]++
			delete(co.activeTasks[cid], tid)
		}
	}
	delete(co.currentView.TaskAssignments, tid)
	co.currentView.ViewNum++
	// Add it to the map of killed tasks so we don't assign it again 
	co.killedTasks[tid] = true
}

func (co *Coordinator) Finish(outputTasks []TaskID) { 
	// When we're completely done 
	co.isFinished = true
	co.outputTasks = outputTasks

	co.currentView.ViewNum++
	co.currentView.isFinished = true
	co.currentView.outputTasks = outputTasks
}  


// For testing purposes 
func (co *Coordinator) Kill() { 
	co.dead = true
	co.l.Close()
	co.px.Kill()
} 

// So that we can start ticks and block if necessary 
func (co *Coordinator) StartTicks() {
	for !co.dead {
		co.tick()
		time.Sleep(250 * time.Millisecond)
	}
}

func StartServer(servers []string, me int, dc DeveloperCoord, numTaskReplicas int, seed int64, socktype string) *Coordinator { 
	co := MakeServer(servers, me, dc, numTaskReplicas, seed, socktype)
	go co.StartTicks()
	return co
}


func MakeServer(servers []string, me int, dc DeveloperCoord, numTaskReplicas int, seed int64, socktype string) *Coordinator { 
	gob.Register(Op{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	co := new(Coordinator)
	co.socktype = socktype
	co.me = me 
	co.dc = dc 
	co.currentView = View{ 0, false, nil, map[TaskID]TaskParams{}, map[TaskID][]ClientID{}, map[TaskID][]ClientID{}, map[TaskID]TaskInfo{}, map[ClientID]string{}}
	co.initialized = false
	co.leaderID = 0
	co.leaderNum = 0	
	co.currentSeq = -1
	co.lastQueries = map[ClientID]int{}
	co.lastLeaderElection = time.Now()

	co.seed = seed
	co.numTaskReplicas = numTaskReplicas
	co.unassignedTasks = list.New()
	co.activeTasks = map[ClientID]map[TaskID]bool{}
	co.availableClients = map[ClientID]int{}
	co.killedTasks = map[TaskID]bool{}
	co.finishedTasks = map[TaskID]bool{}

	co.isFinished = false

	rpcs := rpc.NewServer()
	rpcs.Register(co)

	co.px = paxos.Make(servers, me, rpcs, socktype)

	// TODO: change this implementation for TCP sockets when necessary 
	os.Remove(servers[me])
	l, e := net.Listen(co.socktype, servers[me])
	if e != nil { 
		log.Fatal("listen error: ", e)
	} 
	co.l = l

	// Code to listen and serve requests
	go func() { 
		for !co.dead { 
			conn, err := co.l.Accept()
			if err == nil && !co.dead { 
				if co.unreliable && (rand.Int63() % 1000) < 100 { 
					// discard the request 
					conn.Close()
				} else if co.unreliable && (rand.Int63() % 1000) < 200 { 
					// process request but discard reply 
					c1 := conn.(*net.UnixConn)
					f, _ := c1.File()
					err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
					if err != nil { 
						fmt.Printf("shutdown: %v\n", err)
					}
					go rpcs.ServeConn(conn)
				} else { 
					go rpcs.ServeConn(conn)
				}  
			} else if err == nil { 
				conn.Close()
			}
			if err != nil && !co.dead { 
				fmt.Printf("Coordinator(%v) accept: %v\n", me, err.Error())
				co.Kill()
			} 
		} 
	}() 

	return co
} 

