package main

import "coord"
import "fmt"
import "time"
import "net"
import "os"
import "log"
import "net/rpc"
import "runtime"
import "flag"
import "strings"

type Options struct {
  servers []string
  socket string
}

type Client struct { 
  id coordinator.ClientID
  l net.Listener
  clerk *coordinator.Clerk
  currentView coordinator.View
  options Options
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

func InitFlags() *Options {
  options := new(Options)
  servers := flag.String("servers", "", "comma-seperated list of servers")
  socket := flag.String("socket", "", "name of the client-dev socket")

  flag.Parse()

  if *servers == "" || *socket == "" {
    log.Fatal("usage: -servers ip:port[,ip:port...] -socket path")
  } 

  options.servers = strings.Split(*servers, ",")
  options.socket = *socket
  return options
}

func Init(opts *Options) *Client {
  c := new(Client)
  c.clerk = coordinator.MakeClerk(opts.servers, opts.socket, c.numNodes())

  go func() {
    for {
      c.tick()
      time.Sleep(250 * time.Millisecond)
    }
  }()

  c.startServer(opts.socket)
  return c
}

func main() {
  opts := InitFlags()
  Init(opts)
}
