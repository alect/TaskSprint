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


type NodeSocket string
type Client struct { 
  viewMu sync.Mutex
  id coordinator.ClientID
  l net.Listener
  clerk *coordinator.Clerk
  currentView coordinator.View
  options *Options
  nodes []NodeSocket
  tasks []coordinator.TaskID
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

// Wow, ridiculously ugly.
func (c *Client) ExtractTasks(view *coordinator.View) []coordinator.TaskID {
  tasks := make([]coordinator.TaskID, 0)
  for tid, v := range view.TaskAssignments {
    for _, cid := range v {
      if c.id == cid { tasks = append(tasks, tid) }
    }
  }
  return tasks
}

func (c *Client) SplitTasks(
tasks []coordinator.TaskID) ([]coordinator.TaskID, []coordinator.TaskID) {
  currentTasks := make(map[coordinator.TaskID]int)
  for _, tid := range c.tasks { currentTasks[tid] = 1 }

  newTasks := make([]coordinator.TaskID, 0)
  for _, tid := range tasks {
    _, present := currentTasks[tid]
    if !present { newTasks = append(newTasks, tid) }
    currentTasks[tid] = 0
  }

  killedTasks := make([]coordinator.TaskID, 0)
  for k, v := range currentTasks {
    if v == 1 && k != -1 { killedTasks = append(killedTasks, k) }
  }

  return newTasks, killedTasks
}

func (c *Client) killTasks(tasks []coordinator.TaskID) {
  fmt.Printf("Killing %v\n", tasks)

}

func (c *Client) scheduleTasks(tasks []coordinator.TaskID, 
args map[coordinator.TaskID]coordinator.TaskParams) {
  fmt.Printf("Scheduling %v\n", tasks)
  t := 0
  for i := 0; i < len(c.tasks) && t < len(tasks); i, t = i + 1, t + 1 {
    if c.tasks[i] == -1 {
      c.tasks[i] = tasks[t]
      go c.runTask(i, args[tasks[t]])
    }
  }
}

func (c *Client) runTask(index int, params coordinator.TaskParams) {
  fmt.Printf("Running task %d on %s\n", c.tasks[index], c.nodes[index])
  fmt.Printf("Params: %v\n", params)
}

func (c *Client) processView(view *coordinator.View) {
  PrintView(view)
  c.viewMu.Lock()
  defer c.viewMu.Unlock()

  myTasks := c.ExtractTasks(view)
  newTasks, killedTasks := c.SplitTasks(myTasks)
  c.killTasks(killedTasks)
  c.scheduleTasks(newTasks, view.TaskParams)

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
  c.nodes = make([]NodeSocket, cpus)
  c.tasks = make([]coordinator.TaskID, cpus)
  for i := 0; i < cpus; i++ {
    socket := "/tmp/ts-client-node-" + strconv.Itoa(i)
    c.nodes[i] = NodeSocket(socket)
    c.tasks[i] = -1

    // Starting the subproc and copying its stdout to mine
    cmd := exec.Command(c.options.program, socket)
    stdout, outerr := cmd.StdoutPipe()
    if outerr != nil { log.Fatal(outerr) }
    if err := cmd.Start(); err != nil { log.Fatal(err) }
    go io.Copy(os.Stdout, stdout)

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

func Init(o *Options) *Client {
  c := new(Client)
  c.options = o
  c.id = coordinator.ClientID(nrand())
  c.clerk = coordinator.MakeClerk(o.servers, o.socket, c.initNodes(), c.id)
  c.currentView.ViewNum = -1

  go func() {
    for {
      c.tick()
      time.Sleep(250 * time.Millisecond)
    }
  }()

  c.startServer(o.socket)
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
