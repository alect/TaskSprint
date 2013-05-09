package paxos

//
// Paxos library, to be included in an application.
// Multiple applications will run, each including
// a Paxos peer.
//
// Manages a sequence of agreed-on values.
// The set of peers is fixed.
// Copes with network failures (partition, msg loss, &c).
// Does not store anything persistently, so cannot handle crash+restart.
//
// The application interface:
//
// px = paxos.Make(peers []string, me string)
// px.Start(seq int, v interface{}) -- start agreement on new instance
// px.Status(seq int) (decided bool, v interface{}) -- get info about an instance
// px.Done(seq int) -- ok to forget all instances <= seq
// px.Max() int -- highest instance seq known, or -1
// px.Min() int -- instances before this seq have been forgotten
//

import "net"
import "net/rpc"
import "log"
import "os"
import "syscall"
import "sync"
import "fmt"
import "math/rand"
import "time"


type Paxos struct {
  mu sync.Mutex
  l net.Listener
  dead bool
  unreliable bool
  rpcCount int
  peers []string
  me int // index into peers[]


  // Your data here.
	agreementInsts map[int]PaxosInstance
	maxSeq int
	// The max done value we've seen 
	maxDone int 
	// The max dones from each of our peers 
	dones []int

	// The socket type for rpcs
	socktype string
}

type PaxosInstance struct { 
	n_p int // Highest prepare seen 
	n_a int // Highest accept seen 
	v_a interface{} // Value corresponding to highest accept
	decided bool
	proposing bool // Whether this peer has started a proposer
} 

type PrepareArgs struct { 
	Seq int // The instance sequence number 
	N int // The proposal number
	// For garbage collection 
	Done int 
	Me int
} 

type PrepareReply struct { 
	OK bool // Whether the prepare was accepted 
	N_P int // If we reject the prepare, the highest prepare we've seen so far
	N_A int // The highest proposal accepted so far
	V_A interface{} // The value associated with the highest proposal
} 

type AcceptArgs struct { 
	Seq int // The instance sequence number 
	N int // The proposal number 
	V interface{} // The value associated with the proposal
	// For garbage collection 
	Done int 
	Me int
} 

type AcceptReply struct { 
	OK bool
	N_P int // If we reject the accept, the highest prepare we've seen so far
} 

type DecideArgs struct { 
	Seq int
	N int
	V interface{}
	Done int
	Me int
} 

type DecideReply struct { 
} 




//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the replys contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it does not get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please do not change this function.
//
func call(srv string, name string, args interface{}, reply interface{}, socktype string) bool {
  c, err := rpc.Dial(socktype, srv)
  if err != nil {
    err1 := err.(*net.OpError)
    if err1.Err != syscall.ENOENT && err1.Err != syscall.ECONNREFUSED {
      fmt.Printf("paxos Dial() failed: %v\n", err1)
    }
    return false
  }
  defer c.Close()
    
  err = c.Call(name, args, reply)
  if err == nil {
    return true
  }
  return false
}


func (px *Paxos) HandleDones(peer int, done int) { 
	if done > px.dones[peer] { 
		px.dones[peer] = done
		px.GarbageCollect()
	} 
} 

func (px *Paxos) GarbageCollect() { 
	min := px.Min()
	for k, _ := range px.agreementInsts { 
		if k < min { 
			delete(px.agreementInsts, k)
		} 
	} 
} 

// For the acceptor, the prepare function
func (px *Paxos) Prepare(args *PrepareArgs, reply *PrepareReply) error { 
	px.mu.Lock()
	defer px.mu.Unlock() 

	px.HandleDones(args.Me, args.Done)

	if args.Seq > px.maxSeq { 
		px.maxSeq = args.Seq
	} 

	// First, determine the paxos instance we're hearing about
	pxState, exists := px.agreementInsts[args.Seq]
	if !exists { 
		pxState = PaxosInstance{-1, -1, nil, false, false}
	} 
	if args.N > pxState.n_p { 
		pxState.n_p = args.N
		reply.OK = true
		reply.N_A = pxState.n_a
		reply.V_A = pxState.v_a
	} else { 
		reply.OK = false
		reply.N_P = pxState.n_p
	} 
	px.agreementInsts[args.Seq] = pxState
	return nil
}

// For the acceptor, the accept function 
func (px *Paxos) Accept(args *AcceptArgs, reply *AcceptReply) error { 
	px.mu.Lock()
	defer px.mu.Unlock() 

	px.HandleDones(args.Me, args.Done)

	if args.Seq > px.maxSeq { 
		px.maxSeq = args.Seq
	} 

	pxState, exists := px.agreementInsts[args.Seq] 
	if !exists { 
		pxState = PaxosInstance{-1, -1, nil, false, false} 
	} 
	if args.N >= pxState.n_p { 
		pxState.n_p = args.N 
		pxState.n_a = args.N
		pxState.v_a = args.V
		reply.OK = true
	} else { 
		reply.OK = false
		reply.N_P = pxState.n_p
	}

	px.agreementInsts[args.Seq] = pxState
	return nil
} 

// For the acceptor, when a value has been decided for an instance 
func (px *Paxos) Decide(args *DecideArgs, reply *DecideReply) error { 
	px.mu.Lock()
	defer px.mu.Unlock()
	px.HandleDones(args.Me, args.Done)
	px.agreementInsts[args.Seq] = PaxosInstance{args.N, args.N, args.V, true, true}
	return nil
} 

// The proposer sub-routine
func (px *Paxos) Propose(seq int, v interface{}) { 
	var prepareArgs PrepareArgs 
	var acceptArgs AcceptArgs 

	// Keeping track of the highest N we've seen from replies 
	nMax := -1
	// Keep proposing until a decision is reached
	for !px.dead { 
		// First, decide what proposal number to use. Should be higher than any we've been made aware of 
		// Start by peeking at our highest accepted number 
		state := px.agreementInsts[seq]
		if state.decided { 
			return
		} 
		n_a := state.n_a
		v_a := state.v_a
		n := n_a+1
		if state.n_p >= n { 
			n = state.n_p+1
		} 
		if nMax >= n { 
			n = nMax+1
		} 
		// Start by sending prepare messages to all the peers
		numPrepared := 0
		prepareArgs = PrepareArgs{ seq, n, px.maxDone, px.me }
		//fmt.Printf("Server %v sending prepares for %v\n", px.me, n); 
		for i := 0; i < len(px.peers) && !px.dead; i++ { 
			var prepareReply PrepareReply
			ok := call(px.peers[i], "Paxos.Prepare", prepareArgs, &prepareReply, px.socktype)
			if ok {
				if prepareReply.OK { 
					numPrepared++
					if prepareReply.N_A > n_a { 
						n_a = prepareReply.N_A
						v_a = prepareReply.V_A
					} 
				} else if prepareReply.N_P > nMax { 
					nMax = prepareReply.N_P
				} 
			} 
		}
		// If we have enough preparers, then begin accepting 
		if numPrepared > len(px.peers)/2 { 
			if v_a != nil {
				v = v_a
			} 
			numAccepted := 0 
			acceptArgs = AcceptArgs{ seq, n, v, px.maxDone, px.me } 
			for i := 0; i < len(px.peers) && !px.dead; i++ { 
				var acceptReply AcceptReply
				ok := call(px.peers[i], "Paxos.Accept", acceptArgs, &acceptReply, px.socktype)
				if ok { 
					if acceptReply.OK { 
						numAccepted++ 
					} 
					if acceptReply.N_P > nMax { 
						nMax = acceptReply.N_P
					} 
				} 
			}
			// If we have enough acceptors, it's decided 
			if numAccepted > len(px.peers)/2 {
				decidedArgs := DecideArgs { seq, n, v, px.maxDone, px.me }
				var decidedReply DecideReply
				for i := 0; i < len(px.peers) && !px.dead; i++ { 
					call(px.peers[i], "Paxos.Decide", decidedArgs, &decidedReply, px.socktype)
				}
				px.Decide(&decidedArgs, &decidedReply)
				return
			} 
		} 
		// sleep between proposal attempts 
		time.Sleep(time.Millisecond*10)
	} 
} 



//
// the application wants paxos to start agreement on
// instance seq, with proposed value v.
// Start() returns right away; the application will
// call Status() to find out if/when agreement
// is reached.
//
func (px *Paxos) Start(seq int, v interface{}) {
  // Your code here.
	px.mu.Lock()
	if seq < px.Min() {
		px.mu.Unlock()
		return
	} 
	if seq > px.maxSeq { 
		px.maxSeq = seq
	} 
	state, exists := px.agreementInsts[seq]
	// If we're already handling the instance and we haven't accepted anything yet, 
	// don't start a new instance 
	shouldStart := true
	if exists { 
		shouldStart = !state.proposing
	} else { 
		state = PaxosInstance{-1, -1, nil, false, true}
		px.agreementInsts[seq] = state
	}
	if shouldStart { 
		go px.Propose(seq, v)
	}
	px.mu.Unlock()
}

//
// the application on this machine is done with
// all instances <= seq.
//
// see the comments for Min() for more explanation.
//
func (px *Paxos) Done(seq int) {
  // Your code here.
	px.mu.Lock()
	if seq > px.maxDone { 
		px.maxDone = seq
		px.HandleDones(px.me, px.maxDone)
	} 
	px.mu.Unlock()
}

//
// the application wants to know the
// highest instance sequence known to
// this peer.
//
func (px *Paxos) Max() int {
  // Your code here.
  return px.maxSeq
}

//
// Min() should return one more than the minimum among z_i,
// where z_i is the highest number ever passed
// to Done() on peer i. A peers z_i is -1 if it has
// never called Done().
//
// Paxos is required to have forgotten all information
// about any instances it knows that are < Min().
// The point is to free up memory in long-running
// Paxos-based servers.
//
// It is illegal to call Done(i) on a peer and
// then call Start(j) on that peer for any j <= i.
//
// Paxos peers need to exchange their highest Done()
// arguments in order to implement Min(). These
// exchanges can be piggybacked on ordinary Paxos
// agreement protocol messages, so it is OK if one
// peers Min does not reflect another Peers Done()
// until after the next instance is agreed to.
//
// The fact that Min() is defined as a minimum over
// *all* Paxos peers means that Min() cannot increase until
// all peers have been heard from. So if a peer is dead
// or unreachable, other peers Min()s will not increase
// even if all reachable peers call Done. The reason for
// this is that when the unreachable peer comes back to
// life, it will need to catch up on instances that it
// missed -- the other peers therefor cannot forget these
// instances.
// 
func (px *Paxos) Min() int {
  // You code here.
	// Go through our dones and return the min 

	min := px.dones[px.me]
	for _, peerDone := range px.dones { 
		if peerDone < min { 
			min = peerDone
		} 
	} 

  return min+1
}

func (px *Paxos) MaxDone() int { 
	return px.maxDone+1
} 

//
// the application wants to know whether this
// peer thinks an instance has been decided,
// and if so what the agreed value is. Status()
// should just inspect the local peer state;
// it should not contact other Paxos peers.
//
func (px *Paxos) Status(seq int) (bool, interface{}) {
  // Your code here.
	px.mu.Lock()
	defer px.mu.Unlock()
	if seq < px.Min() { 
		return false, nil
	} 

	state, exists := px.agreementInsts[seq]	
	if exists { 
		return state.decided, state.v_a
	}
	return false, nil
}


//
// tell the peer to shut itself down.
// for testing.
// please do not change this function.
//
func (px *Paxos) Kill() {
  px.dead = true
  if px.l != nil {
    px.l.Close()
  }
}

//
// the application wants to create a paxos peer.
// the ports of all the paxos peers (including this one)
// are in peers[]. this servers port is peers[me].
//
func Make(peers []string, me int, rpcs *rpc.Server, socktype string) *Paxos {
  px := &Paxos{}
  px.peers = peers
  px.me = me
	px.agreementInsts = map[int]PaxosInstance{}
	px.maxSeq = -1
	px.maxDone = -1
	px.dones = make([]int, len(peers))
	px.socktype = socktype
	for i := range px.dones { 
		px.dones[i] = -1
	} 
	

  // Your initialization code here.
		
  if rpcs != nil {
    // caller will create socket &c
    rpcs.Register(px)
  } else {
    rpcs = rpc.NewServer()
    rpcs.Register(px)

    // prepare to receive connections from clients.
    // change "unix" to "tcp" to use over a network.
    os.Remove(peers[me]) // only needed for "unix"
    l, e := net.Listen(socktype, peers[me]);
    if e != nil {
      log.Fatal("listen error: ", e);
    }
    px.l = l
    
	
    // please do not change any of the following code,
    // or do anything to subvert it.
    
    // create a thread to accept RPC connections
    go func() {
      for px.dead == false {
				
				conn, err := px.l.Accept()
        if err == nil && px.dead == false {
          if px.unreliable && (rand.Int63() % 1000) < 100 {
            // discard the request.
            conn.Close()
          } else if px.unreliable && (rand.Int63() % 1000) < 200 {
            // process the request but force discard of reply.
            c1 := conn.(*net.UnixConn)
            f, _ := c1.File()
            err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
            if err != nil {
              fmt.Printf("shutdown: %v\n", err)
            }
            px.rpcCount++
            go rpcs.ServeConn(conn)
          } else {
            px.rpcCount++
            go rpcs.ServeConn(conn)
          }
        } else if err == nil {
          conn.Close()
        }
        if err != nil && px.dead == false {
          fmt.Printf("Paxos(%v) accept: %v\n", me, err.Error())
        }
      }
    }()
  }


  return px
}
