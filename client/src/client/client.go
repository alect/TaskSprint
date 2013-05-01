package main

import "coord"
import "fmt"
import "time"
import "net"
import "os"
import "log"
import "net/rpc"
import "runtime"

type Client struct { 
  id coordinator.ClientID
  l net.Listener
  clerk *coordinator.Clerk
  currentView coordinator.View
}

type GetDataArgs struct { 
	Key string 
} 

type GetDataReply struct { 
	Data interface{}
} 

func (c *Client) GetData(args *GetDataArgs, reply *GetDataReply) error {
  return nil
}

func (c *Client) printView(view *coordinator.View) {
  fmt.Printf("%v\n", view)
}

func (c *Client) tick() {
  latestView := c.clerk.Query()
  c.printView(&latestView)
}


func (c *Client) startServer(socket string) {
  rpcs := rpc.NewServer()
  rpcs.Register(c)

  os.Remove(socket) // For now...
  l, e := net.Listen("unix", socket);
  if e != nil {
    log.Fatal("listen error: ", e);
  }
  c.l = l

  for {
    conn, err := c.l.Accept()
    if err != nil {
      fmt.Printf("RPC Error: %v\n", err.Error())
      c.l.Close()
    } else {
      go rpcs.ServeConn(conn)
    }
  }
}

func (c *Client) numNodes() int {
  return runtime.NumCPU();
}

func Init(servers[]string, socket string) *Client {
  c := new(Client)
  c.clerk = coordinator.MakeClerk(servers, socket, c.numNodes())

  go func() {
    for {
      c.tick()
      time.Sleep(250 * time.Millisecond)
    }
  }()

  c.startServer(socket)
  return c
}

func main() {
  servers := make([]string, 0)
  Init(servers, "/tmp/socket")
}
