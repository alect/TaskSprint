package coordinator 

import "net/rpc"
import "time"

type Clerk struct { 
	servers []string // Coordinator replicas 
}

func MakeClerk(servers[]string) *Clerk { 
	ck := new(Clerk)
	ck.servers = servers 
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
			args := &QueryArgs{}
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

