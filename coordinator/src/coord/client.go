package coordinator 

import "net/rpc"
import "time"

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
	for { 
		// try each known server. 
		for _, srv := range ck.servers { 
			args := &QueryArgs{ CID: ck.clerkID, Contact: ck.me, NumNodes: ck.numNodes }
			var reply QueryReply
			ok := call(srv, "Coordinator.Query", args, &reply, ck.socktype)
			if ok { 
				return reply.View
			} 
		}
		time.Sleep(100 * time.Millisecond)
	} 
	return View{}
} 

func (ck *Clerk) Done(TID TaskID, DoneValues map[string]interface{}) { 
	for {
		for _, srv := range ck.servers {
			args := &DoneArgs { ck.clerkID, TID, DoneValues}
			var reply DoneReply
			ok := call(srv, "Coordinator.TaskDone", args, &reply, ck.socktype)
			if ok {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
