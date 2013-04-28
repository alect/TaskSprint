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

