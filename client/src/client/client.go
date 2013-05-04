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
import "encoding/json"

type Options struct {
  servers []string
  socket string
  program string
}

type NodeStatus int
const (
  Free = 0
  Busy = 1
)

type TaskStatus int
const (
  Pending = 0
  Started = 1
  Finished = 2
  Killed = 3
)

type NodeID int
type Node struct {
  socket string
  status NodeStatus
  task *Task
}

type Task struct {
  id coordinator.TaskID
  node *Node
  status TaskStatus
  params coordinator.TaskParams
  data map[string]interface{}
}

type Client struct {
  viewMu sync.Mutex
  id coordinator.ClientID
  l net.Listener
  clerk *coordinator.Clerk
  currentView coordinator.View
  options *Options
  nodes []*Node
  tasks map[coordinator.TaskID]*Task
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
    buffer.WriteString("\n")
  }
  fmt.Println(buffer.String())
}

func (c *Client) processView(view *coordinator.View) {
  /* PrintView(view) */
  c.viewMu.Lock()
  defer c.viewMu.Unlock()

  myTasks := c.ExtractTasks(view)
  newTasks, killedTasks := c.SplitTasks(myTasks)
  c.killTasks(killedTasks)
  c.scheduleTasks(newTasks, view.TaskParams)

  c.currentView = *view;
}

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
  // Do I need a task lock? Not sure.
  currentTasks := make(map[coordinator.TaskID]int)
  for tid, _ := range c.tasks { currentTasks[tid] = 1 }

  newTasks := make([]coordinator.TaskID, 0)
  for _, tid := range tasks {
    _, present := currentTasks[tid]
    if !present { newTasks = append(newTasks, tid) }
    currentTasks[tid] = 0
  }

  killedTasks := make([]coordinator.TaskID, 0)
  for k, v := range currentTasks {
    if v == 1 { killedTasks = append(killedTasks, k) }
  }

  return newTasks, killedTasks
}

func (c *Client) killTasks(tasks []coordinator.TaskID) {
  // Need to lock task array
  /* fmt.Printf("Killing %v\n", tasks) */
  for _, tid := range tasks {
    c.tasks[tid].status = Killed
  }
}

func (c *Client) scheduleTasks(tasks []coordinator.TaskID,
args map[coordinator.TaskID]coordinator.TaskParams) {
  // Need to lock task array
  /* fmt.Printf("Scheduling %v\n", tasks) */

	/*
	t := 0
  for i := 0; i < len(c.nodes) && t < len(tasks); i, t = i + 1, t + 1 {
    if c.nodes[i].status == Free {
      newTask := &Task{tasks[t], c.nodes[i], Pending, args[tasks[t]], nil}
      c.tasks[tasks[t]] = newTask
      go c.runTask(newTask)
    }
  }
	*/
	// Try to schedule each task on a free node
	for t := 0; t < len(tasks); t++ { 
		for i := 0; i < len(c.nodes); i++ { 
			if c.nodes[i].status == Free { 
				newTask := &Task{tasks[t], c.nodes[i], Pending, args[tasks[t]], nil}
				// Have to make the node busy here
				c.nodes[i].status = Busy
				c.tasks[tasks[t]] = newTask
				go c.runTask(newTask)
				break
			} 
		}
	} 
}

func (c *Client) runTask(task *Task) {
  // Need task lock
  node, params := task.node, task.params
  //fmt.Printf("Running task %d on %s\n", task.id, node.socket)
  /* fmt.Printf("params: %v\n", params) */

  // Connecting to node and marking task as started
  node.status, task.status = Busy, Started
  conn, err := net.Dial("unix", node.socket)
  if err != nil { log.Fatal("Node is dead.", err) }

  // Sending data 
  data := fmt.Sprintf("[\"%s\", %s]", params.FuncName, params.BaseObject)
  fmt.Fprintf(conn, data)

  // Waiting for result
  buffer := make([]byte, 1024)
  size, readerr := conn.Read(buffer)
  if readerr != nil { log.Fatal(readerr) }
  conn.Close()

  // Unserializing and marking as finished
  result := make(map[string]interface{})
  parseerr := json.Unmarshal(buffer[:size], &result)
  if parseerr != nil { log.Fatal(parseerr) }
  c.markFinished(task, result)
}

func (c *Client) markFinished(task *Task, result map[string]interface{}) {
  // Need task lock
  /* fmt.Printf("Result is %v\n", result) */
  outResult := make(map[string]interface{})
  for _, k := range task.params.DoneKeys {
    if v, p := result[k]; p { outResult[k] = v }
  }

  c.clerk.Done(task.id, outResult)
  node := task.node
  node.task = nil
  node.status = Free

  task.data = result
  task.status = Finished
  task.node = nil
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
  cpus := runtime.NumCPU();
  c.nodes = make([]*Node, cpus)
  for i := 0; i < cpus; i++ {
    socket := "/tmp/ts-client-node-"
    socket += strconv.FormatInt(int64(c.id), 10) + "-" + strconv.Itoa(i)
    c.nodes[i] = &Node{socket, Free, nil}
    // Starting the subproc and copying its stdout to mine
    cmd := exec.Command(c.options.program, socket)
    stdout, outerr := cmd.StdoutPipe()
    if outerr != nil { log.Fatal(outerr) }
    if err := cmd.Start(); err != nil { log.Fatal(err) }
    go io.Copy(os.Stdout, stdout)

    time.Sleep(250 * time.Millisecond)
    fmt.Printf("Node at %s started.\n", socket)
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
	//c.viewMu.Lock()

  c.options = o
  c.id = coordinator.ClientID(nrand())
  c.clerk = coordinator.MakeClerk(o.servers, o.socket, c.initNodes(), c.id)
  c.tasks = make(map[coordinator.TaskID]*Task)
  c.currentView.ViewNum = -1

	//c.viewMu.Unlock()

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

