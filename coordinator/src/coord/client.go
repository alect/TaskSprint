package coordinator 

import "net/rpc"
import "time"
import "math/rand"

type Clerk struct { 
	servers []string // Coordinator replicas 
	string me // How to contact this client 
	clerkID ClientID 
}

func MakeClerk(servers[]string, me string) *Clerk { 
	ck := new(Clerk)
	ck.servers = servers 
	ck.me = me 
	ck.clerkID = rand.Int63()
	return ck
} 


// RPC call function 
func call(srv string, rpcname string, args interface{}, reply interface{}) bool { 
	c, errx := rpc.Dial("unix", srv)
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
			args := &QueryArgs{ CID: ck.clerkID, Contact: ck.me }
			var reply QueryReply
			ok := call(srv, "Coordinator.Query", args, &reply)
			if ok { 
				return reply.View
			} 
		}
		time.Sleep(100 * time.Millisecond)
	} 
	return View{}
} 

func (ck *Clerk) Done(TID TaskID, DoneValues map[string]interface) { 
	for { 
		for _, srv := range ck.servers { 
			args := &DoneArgs { ck.clerkID, TID, DoneValues} 
			var reply DoneReply
			ok := call(srv, "Coordinator.Done", args, &reply) 
			if ok { 
				return
			} 
		} 
		time.Sleep(100 * time.Millisecond)
	} 
} 

