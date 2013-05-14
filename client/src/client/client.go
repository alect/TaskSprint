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
import "encoding/gob"

type Options struct {
  servers []string
  socket string
  socktype string
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
  dead bool
}

type GetDataArgs struct {
	Key string
  TaskId coordinator.TaskID
}

type GetDataReply struct {
	Data interface{}
}

func (c *Client) GetData(args *GetDataArgs, reply *GetDataReply) error {
  id, key := args.TaskId, args.Key
  reply.Data = c.tasks[id].data[key]
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

  c.currentView = *view;
  //myTasksNew := c.ExtractTasksNew(view)
  myTasks := c.ExtractTasks(view)

  //fmt.Printf("New: %v\n", myTasksNew)
  //fmt.Printf("Old: %v\n\n", myTasks)

  newTasks, killedTasks := c.SplitTasks(myTasks)
  if len(killedTasks) > 0 { c.killTasks(killedTasks) }
  if len(newTasks) > 0 { c.scheduleTasks(newTasks, view.TaskParams) }
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

func (c *Client) ExtractTasksNew(view *coordinator.View) []coordinator.TaskID {
  tasks := make([]coordinator.TaskID, 0)
  for tid, v := range view.Tasks {
    for _, cid := range v.PendingClients {
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
    if v == 1 {
      task, present := c.tasks[k]
      if !present { log.Fatal("Cannot kill nonexistant task.") }
      if task.status == Killed || task.status == Finished { continue }
      killedTasks = append(killedTasks, k)
    }
  }

  return newTasks, killedTasks
}

func (c *Client) killTasks(tasks []coordinator.TaskID) {
  // Need to lock task array
  if len(tasks) > 0 {
    fmt.Printf("Killing %v\n", tasks)
  }
  for _, tid := range tasks {
    task := c.tasks[tid]
    node := task.node

    c.killNode(node)
    task.node = nil
    task.status = Killed

    node.task = nil
    node.status = Free
    c.startNode(node)
  }
}

func (c *Client) scheduleTasks(tasks []coordinator.TaskID,
args map[coordinator.TaskID]coordinator.TaskParams) {
  /* fmt.Printf("Scheduling %v\n", tasks) */
  t := 0
  for i := 0; i < len(c.nodes) && t < len(tasks); i++ {
    if c.nodes[i].status == Free {
      newTask := &Task{tasks[t], c.nodes[i], Pending, args[tasks[t]], nil}
      c.tasks[tasks[t]] = newTask
			c.nodes[i].status = Busy
      go c.runTask(newTask)
      t++
    }
  }
}

func (c *Client) FetchData(cid coordinator.ClientID,
tid coordinator.TaskID, key string) (interface{}, bool) {
  args := &GetDataArgs{key, tid}
	var reply GetDataReply
  var ok bool
  if cid == c.id {
    c.GetData(args, &reply)
    ok = true
  } else {
    srv := c.currentView.ClientInfo[cid]
    ok = c.call(srv, "Client.GetData", args, &reply)
  }

  return reply.Data, ok && reply.Data != nil
}

func (c *Client) getJson(task string, data interface{}) string {
  jsonString := ""
  if data != nil {
    byteArray, err := json.Marshal(data)
    if err != nil { log.Fatal("Failed to marshal data.") }
    jsonString = string(byteArray)
  }

  return fmt.Sprintf("[\"%s\", %s]", task, jsonString)
}

func (c *Client) fetchParams(params *coordinator.TaskParams) string {
  if len(params.PreReqTasks) == 0 {
    if params.BaseObject == nil { params.BaseObject = [0]int{}}
    return c.getJson(params.FuncName, params.BaseObject)
  }
  prereqs, complete := len(params.PreReqTasks), 0
  fetched := make([]bool, prereqs)
  data := make([]interface{}, prereqs)
  for complete < prereqs && !c.dead {
    for t, done := range fetched {
      if done { continue }
      if c.dead { break }
      tid := params.PreReqTasks[t]
      finishedClients := c.currentView.Tasks[tid].FinishedClients
      for _, cid := range finishedClients {
        //fmt.Printf("Trying %d\n", cid)
        datum, ok := c.FetchData(cid, tid, params.PreReqKey[t])
        if !ok { continue }
        fetched[t], data[t] = true, datum
        complete++
        //fmt.Printf("Got %v\n", datum)
      }
    }

    // Pause before retrying
    time.Sleep(250 * time.Millisecond)
  }

  return c.getJson(params.FuncName, data)
}

func (c *Client) runTask(task *Task) {
  /* fmt.Printf("Starting %v\n", task) */
  node, params := task.node, task.params
  data := c.fetchParams(&params)
  /* fmt.Printf("Running %v\n", task) */

  // Trying to connect to node
  node.status, task.status = Busy, Started
  conn, err := net.Dial("unix", node.socket)
  for tries := 0; tries < 20 && err != nil; tries++ {
    if c.dead { return }
    time.Sleep(250 * time.Millisecond)
    conn, err = net.Dial("unix", node.socket)
  }

  // Unsuccessful; die
  if err != nil { log.Fatal("Node is dead.", err) }

  // Sending data 
  fmt.Fprintf(conn, data)

  // Special kill-task handling
  if task.id == -1 {
    conn.Close()
    return
  }

  // Waiting for result

	buffer := make([]byte, 1024)
  size, readerr := conn.Read(buffer)
  if readerr != nil && readerr != io.EOF {
    log.Fatal("Error reading result. ", readerr)
  }
  conn.Close()
	

  // Unserializing and marking as finished
  result := make(map[string]interface{})
  if size > 0 {
    parseerr := json.Unmarshal(buffer[:size], &result)
    if parseerr != nil { log.Fatal("Error parsing result. ", parseerr) }
  }
  c.markFinished(task, result)
}

func (c *Client) markFinished(task *Task, result map[string]interface{}) {
  outResult := make(map[string]interface{})
  for _, k := range task.params.DoneKeys {
    if v, p := result[k]; p { outResult[k] = v }
  }

  // Need to avoid a very slim, but possible, race
  c.viewMu.Lock()

  node := task.node
  node.task = nil
  node.status = Free

  task.data = result
  task.node = nil
  task.status = Finished

  c.viewMu.Unlock()

  if !c.dead { c.clerk.Done(task.id, outResult) }
}

func (c *Client) tick() bool {
  view := c.clerk.Query()
  if c.currentView.ViewNum < view.ViewNum {
    c.processView(&view)
    return true
  }
  return false
}

func (c *Client) startServer(socket string) {
  rpcs := rpc.NewServer()
  rpcs.Register(c)
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

  if (c.options.socktype == "unix") { os.Remove(socket) }
  l, e := net.Listen(c.options.socktype, socket);
  if e != nil {
    log.Fatal("listen error: ", e);
  }
  c.l = l

  for !c.dead {
    conn, err := c.l.Accept()
    if err != nil {
      fmt.Printf("RPC Error: %v\n", err.Error())
      c.l.Close()
    } else if !c.dead {
      go rpcs.ServeConn(conn)
    } else {
      return
    }
  }
}

func (c *Client) startNode(node *Node) {
  // Starting the subproc and copying its stdout to mine
  cmd := exec.Command(c.options.program, node.socket)

  stdout, outerr := cmd.StdoutPipe()
  if outerr != nil { log.Fatal("Error getting stdout: ", outerr) }
  stderr, errerr := cmd.StderrPipe()
  if errerr != nil { log.Fatal(errerr) }

  if err := cmd.Start(); err != nil { log.Fatal(err) }
  go io.Copy(os.Stdout, stdout)
  go io.Copy(os.Stderr, stderr)
}

func (c *Client) initNodes() int {
  cpus := runtime.NumCPU();
  c.nodes = make([]*Node, cpus)
  for i := 0; i < cpus; i++ {
    socket := c.generateNodeSocket(i)
    node := &Node{socket, Free, nil}
    c.nodes[i] = node
    c.startNode(node)
  }
  fmt.Printf("%d nodes initialized.\n", cpus)
  return cpus
}

func InitFlags() *Options {
  options := new(Options)
  servers := flag.String("servers", "", "comma-seperated list of servers")
  socket := flag.String("socket", "", "name of the client-coord socket")
  program := flag.String("program", "", "path to the dev client executable")
  network := flag.String("network", "", "network type: unix, tcp, udp, etc.")

  flag.Parse()

  if *servers == "" || *socket == "" || *program == "" || *network == "" {
    log.Fatal("usage: -servers ip:port[,ip:port...] " +
      "-socket path -program path -network {unix, tcp}")
  }

  options.servers = strings.Split(*servers, ",")
  options.socket = *socket
	options.socktype = *network
  options.program = *program
  return options
}

func Init(o *Options) *Client {
  c := new(Client)
  c.options = o
  c.id = coordinator.ClientID(nrand())
  c.tasks = make(map[coordinator.TaskID]*Task)
  c.currentView.ViewNum = -1
  c.dead = false
  c.clerk = coordinator.MakeClerk(o.servers, o.socket,
    c.initNodes(), c.id, o.socktype)

  return c
}

// May never return
func (c *Client) Start() {
  go func() {
    for !c.dead {
      if !c.tick() { // Sleep if nothing new
        time.Sleep(100 * time.Millisecond)
      }
    }
  }()

  c.startServer(c.options.socket)
}

func (c *Client) killNode(node *Node) {
  params := new (coordinator.TaskParams)
  params.FuncName = "kill"
  t := &Task{-1, node, Pending, *params, nil}
  c.runTask(t);
}

func (c *Client) Kill() {
  for _, node := range c.nodes {
    c.killNode(node)
  }

  c.dead = true
  c.clearNodeSockets();
}

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
}

func main() {
  Init(InitFlags()).Start()
}

func (c *Client) generateNodeSocket(host int) string {
  s := "/tmp/tsclient-"
  s += strconv.FormatInt(int64(c.id), 10) + "/"
  os.Mkdir(s, 0777)
  s +=  "nodesocket-"
  s += strconv.Itoa(host)
  return s
}

func (c *Client) clearNodeSockets() {
  s := "/tmp/tsclient-"
  s += strconv.FormatInt(int64(c.id), 10) + "/"
  os.RemoveAll(s)
}

// RPC call function 
func (c *Client) call(srv string, rpcname string, args interface{},
	reply interface{}) bool {
  conn, errx := rpc.Dial(c.options.socktype, srv)
  if errx != nil {
    return false
  }
  defer conn.Close()

  err := conn.Call(rpcname, args, reply)
  if err == nil {
    return true
  }
  return false
}
