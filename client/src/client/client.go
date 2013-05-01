package main

import "coord"
import "strconv"
import "fmt"
import "time"
import "net"
import "os"
import "os/exec"
import "io"
import "log"
import "net/rpc"
import "runtime"
import "flag"
import "strings"
import "math/big"
import "crypto/rand"
import "bytes"
import "sync"

type Options struct {
  servers []string
  socket string
  program string
}

type Client struct { 
  viewMu sync.Mutex
  id coordinator.ClientID
  l net.Listener
  clerk *coordinator.Clerk
  currentView coordinator.View
  options *Options
  nodes []string
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

/* TaskParams map[TaskID]TaskParams */
/* TaskAssignments map[TaskID][]ClientID */
/* ClientInfo map[ClientID]string // How to contact a particular client. */
func PrintView(view *coordinator.View) {
  var buffer bytes.Buffer
  fmt.Printf("---------------- View #%d -----------------\n", view.ViewNum)
  for k, v := range view.TaskAssignments {
    args := view.TaskParams[k]
    s := fmt.Sprintf("T%d (%s%v):\t", k, args.FuncName, args.BaseObject)
    buffer.WriteString(s)
    for _, c := range v {
      s := fmt.Sprintf("%d\t", c)
      buffer.WriteString(s)
    }
  }
  fmt.Println(buffer.String())
}

func (c *Client) processView(view *coordinator.View) {
  c.viewMu.Lock()
  defer c.viewMu.Unlock()

  PrintView(view)
  c.currentView = *view;
}

func (c *Client) tick() {
  view := c.clerk.Query()
  if c.currentView.ViewNum < view.ViewNum {
    c.processView(&view)
  }
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

func (c *Client) initNodes() int {
  cpus := runtime.NumCPU() * 2;
  c.nodes = make([]string, cpus)
  for i := 0; i < cpus; i++ {
    socket := "/tmp/ts-client-node-" + strconv.Itoa(i)
    c.nodes[i] = socket

    cmd := exec.Command(c.options.program, socket)
    stdout, outerr := cmd.StdoutPipe()

    if outerr != nil { log.Fatal(outerr) }
    if err := cmd.Start(); err != nil { log.Fatal(err) }
    go io.Copy(os.Stdout, stdout) // Print subproc to stdout

    fmt.Printf("Started node at %s\n", socket)
  }
  return cpus
}

func InitFlags() *Options {
  options := new(Options)
  servers := flag.String("servers", "", "comma-seperated list of servers")
  socket := flag.String("socket", "", "name of the client-coord socket")
  program := flag.String("program", "", "path to the dev client executable")

  flag.Parse()

  if *servers == "" || *socket == "" || *program == "" {
    log.Fatal("usage: -servers ip:port[,ip:port...] -socket path -program path")
  } 

  options.servers = strings.Split(*servers, ",")
  options.socket = *socket
  options.program = *program
  return options
}

func Init(opts *Options) *Client {
  c := new(Client)
  c.options = opts
  c.id = coordinator.ClientID(nrand())
  c.clerk = coordinator.MakeClerk(opts.servers, opts.socket, c.initNodes())
  c.currentView.ViewNum = -1

  go func() {
    for {
      c.tick()
      time.Sleep(250 * time.Millisecond)
    }
  }()

  c.startServer(opts.socket)
  return c
}

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
}

func main() {
  opts := InitFlags()
  Init(opts)
}
