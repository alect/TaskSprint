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
	
	currentView View // The most recent view we're aware of

	isLeader bool // Whether this peer is currently the leader 
	leaderNum int // The current leader number (not a leader id.)
	currentSeq int // The current sequence number (so we don't replay old log entries)
} 

const (
	QUERY = 0
	DONE = 1
	TICK = 2
	LEADER_CHANGE = 3
	NIL = -1
)

type PaxosReply struct { 
	View View
} 

type Op struct {
	// A Paxos log entry 
	Op int
	CID ClientID 
	TID TaskID 
	DoneValues map[string]interface{}
} 

func opsEqual(op1 Op, op2 Op) bool { 
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
		// TODO: See if the client is new and assign tasks as appropriate

		// Clone the current view for queries so concurrency doesn't affect it 
		return cloneView(co.currentView)
	} else if op.Op == DONE { 
		// TODO: Handle finished tasks 
	} else if op.Op == TICK { 
		// TODO: Handle ticks to see if Some clients should be considered dead
	} else if op.Op == LEADER_CHANGE { 
		// TODO: Handle a leader change 
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


// RPC functions 

// When a client wants the latest view
func (co *Coordinator) Query(args *QueryArgs, reply *QueryReply) error { 
	op := Op { Op: QUERY, CID: args.CID }
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



// For testing purposes 
func (co *Coordinator) Kill() { 
	co.dead = true 
	co.l.Close()
	co.px.Kill()
} 


func StartServer(servers []string, me int) *Coordinator { 
	gob.Register(Op{})

	co := new(Coordinator)
	co.me = me 
	
	co.currentSeq = -1

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

	return co
} 