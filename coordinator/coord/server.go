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
	
	isLeader bool // Whether this peer is currently the leader 
	leaderNum int // The current leader number 
	currentSeq int // The current sequence number (so we don't replay old log entries)
} 

type Op struct {
	// A Paxos log entry 
} 


// RPC functions 

// When a client wants the latest view
func (co *Coordinator) Query(args *QueryArgs, reply *QueryReply) error { 

	return nil 
} 

// When a client has finished a task 
func (co *Coordinator) TaskDone(args *DoneArgs, reply *DoneReply) error { 
	
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