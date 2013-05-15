package coordinator 

import "net/rpc"
import "time"
//import "fmt"

type Clerk struct { 
	servers []string // Coordinator replicas 
	me string // How to contact this client 
	clerkID ClientID 
	numNodes int
	socktype string
}

func MakeClerk(servers[]string, me string, numNodes int, id ClientID, socktype string) *Clerk { 
	ck := new(Clerk)
	ck.servers = servers 
	ck.me = me 
	ck.clerkID = id
	ck.numNodes = numNodes
	ck.socktype = socktype
	return ck
} 

// FOR TESTING ONLY
func (ck *Clerk) SetID(cid ClientID) {
	ck.clerkID = cid
}

// RPC call function 
func call(srv string, rpcname string, args interface{}, reply interface{}, socktype string) bool { 
	c, errx := rpc.Dial(socktype, srv)
	if errx != nil { 
		return false
	} 
	defer c.Close()
	
	err := c.Call(rpcname, args, reply)
	if err == nil { 
		return true
	} 
	return false
} 

func (ck *Clerk) Query() View { 
	//fmt.Printf("Client: %v starting query\n", ck.clerkID)
	for { 
		// try each known server. 
		for _, srv := range ck.servers { 
			args := &QueryArgs{ CID: ck.clerkID, Contact: ck.me, NumNodes: ck.numNodes }
			var reply QueryReply
			ok := call(srv, "Coordinator.Query", args, &reply, ck.socktype)
			if ok {
				//fmt.Printf("Client: %v finished query\n", ck.clerkID)
				return reply.View
			} 
		}
		//fmt.Printf("Client: %v query failed\n", ck.clerkID)
		time.Sleep(100 * time.Millisecond)
	} 
	return View{}
} 

func (ck *Clerk) Done(TID TaskID, DoneValues map[string]interface{}) { 
	//fmt.Printf("Client: %v starting done\n", ck.clerkID)
	for {
		for _, srv := range ck.servers {
			args := &DoneArgs { ck.clerkID, TID, DoneValues}
			var reply DoneReply
			ok := call(srv, "Coordinator.TaskDone", args, &reply, ck.socktype)
			if ok {
				//fmt.Printf("Client: %v finished done\n", ck.clerkID)
				return
			}
		}
		//fmt.Printf("Client: %v done failed\n", ck.clerkID)
		time.Sleep(100 * time.Millisecond)
	}
}
