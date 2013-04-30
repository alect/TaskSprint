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

type Coordinator struct { 
	mu sync.Mutex
	l net.Listener
	me int 
	dead bool // For testing 
	unreliable bool // For testing 
	px *paxos.Paxos

	dc *DeveloperCoord // For performing callbacks. This is the developer defined object
	
	currentView View // The most recent view we're aware of
	bool initialized // Whether we've initialized yet. Calls init on first tick if not
	leaderID int // The id of the current leader. == me if we're the leader
	leaderNum int // The current leader number (not a leader id.)
	currentSeq int // The current sequence number (so we don't replay old log entries)

	// How long it's been since we last heard from a client in number of ticks
	lastQueries map[ClientID]int 

	// This state is unique to each replica 
	// Used to determine when replicas will attempt to elect a new leader
	lastLeaderElection time.Time
} 

const (
	QUERY = 0
	DONE = 1
	TICK = 2
	LEADER_CHANGE = 3
	NIL = -1
)

const ELECTION_INTERVAL = 5 * time.Second

const DEAD_TICKS = 10

type PaxosReply struct { 
	View View
} 

type Op struct {
	// A Paxos log entry 
fun	Op int
	CID ClientID 
	Contact string 
	TID TaskID 
	DoneValues map[string]interface{}
	LeaderNum int
	LeaderID int
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
	co.currentSeq = seq 
	if op.Op == QUERY { 
		// See if the client is new and assign tasks as appropriate
		_, exists := co.lastQueries[op.CID] 
		if !exists { 
			// new client event 
			co.dc.ClientJoined(op.CID)
		} 
		co.lastQueries[op.CID] = 0
		// Update our view of how to contact the client 
		co.currentView.ClientInfo[op.CID] = op.Contact

		// Clone the current view for queries so concurrency doesn't affect it 
		return cloneView(co.currentView)
	} else if op.Op == DONE { 
		// Handle finished tasks 
		co.dc.TaskDone(op.TID, op.DoneValues)
	} else if op.Op == TICK { 
		// First, see if we need to initialize everything 
		if !co.initialized { 
			co.dc.Init() 
			co.initialized = true
		} 
		// Handle ticks to see if Some clients should be considered dead
		for clientID, ticks := range co.lastQueries { 
			if ticks+1 == DEAD_TICKS { 
				// Dead client event 
				co.dc.ClientDead(clientID)
			} 
			co.lastQueries[clientID]++
		} 

	} else if op.Op == LEADER_CHANGE { 
		// Handle a leader change 
		// Only change leaders if the new leader is new
		if op.LeaderNum > co.leaderNum { 
			co.leaderNum = op.LeaderNum
			co.leaderID = op.LeaderID
			co.lastLeaderElection = time.Now()
		}
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


// RPC functions 

// When a client wants the latest view
func (co *Coordinator) Query(args *QueryArgs, reply *QueryReply) error { 
	op := Op { Op: QUERY, CID: args.CID, Contact: args.Contact }
	result := co.PerformPaxos(op)
	reply.View = result
	return nil 
} 

// When a client has finished a task 
func (co *Coordinator) TaskDone(args *DoneArgs, reply *DoneReply) error { 
	op := Op { Op: DONE, CID: args.CID, TID: args.TID, DoneValues: args.DoneValues }
	result := co.PerformPaxos(op)
	return nil
} 


// Functions called by the developer coordinator to manage tasks 
func (co *Coordinator) StartTask(params TaskParams) TaskID { 
	// TODO: Implement the function that starts a task 
	return 0
} 

func (co *Coordinator) KillTask(tid TaskID) { 
	// TODO: Implement the killing function 
}

func (co *Coordinator) Finished(outputTasks []TaskID) { 
	// TODO: Implement the finishing function 
}  



// For testing purposes 
func (co *Coordinator) Kill() { 
	co.dead = true 
	co.l.Close()
	co.px.Kill()
} 


func StartServer(servers []string, me int, dc DeveloperCoord) *Coordinator { 
	gob.Register(Op{})

	co := new(Coordinator)
	co.me = me 
	co.dc = dc 
	co.currentView = View{map[TaskID]TaskParams{}, map[TaskID][]ClientID{}, map[ClientID]string{}}
	co.initialized = false
	co.leaderID = 0
	co.leaderNum = 0	
	co.currentSeq = -1
	co.lastQueries = map[ClientID]int
	co.lastLeaderElection = time.Now()

	rpcs := rpc.NewServer()
	rpcs.Register(co)

	co.px = paxos.Make(servers, me, rpcs)

	// TODO: change this implementation for TCP sockets when necessary 
	os.Remove(servers[me])
	l, e := net.Listen("unix", servers[me])
	if e != nil { 
		log.Fatal("listen error: ", e)
	} 
	sm.l = l

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

	go func() { 
		for !co.dead { 
			co.tick()
			time.Sleep(250 * time.Millisecond)
		} 
	} 

	return co
} 