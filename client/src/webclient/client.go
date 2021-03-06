package main

import "coord"
import "fmt"
import "time"
import "net"
import "os"
import "log"
import "net/rpc"
import "flag"
import "strings"
import "math/big"
import "crypto/rand"
import "bytes"
import "sync"
import "encoding/json"
import "encoding/gob"
import "code.google.com/p/go.net/websocket"

type Options struct {
  servers []string
  socket string
  socktype string
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
  socket *websocket.Conn
  status NodeStatus
  task *Task
  taskChannel chan *Task
  finishChannel chan string
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

// FOR TESTING ONLY
func (c *Client) SetID(cid coordinator.ClientID) {
  c.id = cid
  c.clerk.SetID(cid)
}

func (c *Client) GetID() coordinator.ClientID {
  return c.id
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
  /* myTasksNew := c.ExtractTasksNew(view) */
  myTasks := c.ExtractTasks(view)

  /* fmt.Printf("New: %v\n", myTasksNew) */
  /* fmt.Printf("Old: %v\n\n", myTasks) */

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
  // Need to lock task array?
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
        /* fmt.Printf("Trying %d\n", cid) */
        datum, ok := c.FetchData(cid, tid, params.PreReqKey[t])
        if !ok { continue }
        fetched[t], data[t] = true, datum
        complete++
        /* fmt.Printf("Got %v\n", datum) */
      }
    }

    // Pause before retrying
    time.Sleep(250 * time.Millisecond)
  }

  return c.getJson(params.FuncName, data)
}

func (c *Client) runTask(task *Task) {
  /* fmt.Printf("Starting %v\n", task) */
  node := task.node
  /* fmt.Printf("Running %v\n", task) */

  // Sending to node's task channel and waiting for finish
  node.taskChannel <- task
  complete := <- node.finishChannel

  // Special kill-task handling
  if task.id == -1 { return }

  // Client quit. Need to release node.
  if complete == "quit" {
    c.viewMu.Lock()
    log.Fatal("node quit.")
    c.viewMu.Unlock()
    return
  }

  // Deserializing and marking as finished
  result := make(map[string]interface{})
  parseerr := json.Unmarshal([]byte(complete), &result)
  if parseerr != nil { log.Fatal("Error parsing result. ", parseerr) }
  c.markFinished(task, result)
}

func (c *Client) markFinished(task *Task, result map[string]interface{}) {
  outResult := make(map[string]interface{})
  for _, k := range task.params.DoneKeys {
    if v, p := result[k]; p { outResult[k] = v }
  }

  // Need to avoid a very slim, but possible, race
  c.viewMu.Lock()

  if task.status == Killed {
    c.viewMu.Unlock()
    return
  }

  node := task.node
  node.task = nil
  node.status = Free

  task.data = result
  task.node = nil
  task.status = Finished

  c.viewMu.Unlock()

  if !c.dead { c.clerk.Done(task.id, outResult) }
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

func (c *Client) tick() bool {
  // Now need to continously check process the view.
  view := c.clerk.Query()
  c.processView(&view)
  return false

  // We could do this before. No more.
  /* if c.currentView.ViewNum < view.ViewNum { */
    /* c.processView(&view) */
    /* return true */
  /* } */
  /* return false */
}

func (c *Client) startServer(socket string) {
  rpcs := rpc.NewServer()
  rpcs.Register(c)
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

  if (c.options.socktype == "unix") { os.Remove(socket) }
  if (c.options.socktype == "tcp") {
    sepIndex := strings.Index(socket, ":")
    port := socket[sepIndex :]
    socket = "0.0.0.0" + port;
  }
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

func InitFlags() *Options {
  options := new(Options)
  servers := flag.String("servers", "", "comma-seperated list of servers")
  socket := flag.String("socket", "", "name of the client-coord socket")
  network := flag.String("network", "", "network type: unix, tcp, udp, etc.")

  flag.Parse()

  if *servers == "" || *socket == "" || *network == "" {
    log.Fatal("usage: -servers ip:port[,ip:port...] " +
      "-socket path -program path -network {unix, tcp}")
  }

  options.servers = strings.Split(*servers, ",")
  options.socket = *socket
	options.socktype = *network
  return options
}

func Init(o *Options) *Client {
  maxNodes := 1024
  c := new(Client)
  c.options = o
  c.id = coordinator.ClientID(nrand())
  c.tasks = make(map[coordinator.TaskID]*Task)
  c.currentView.ViewNum = -1
  c.dead = false
  c.nodes = make([]*Node, 0)
  c.clerk = coordinator.MakeClerk(o.servers, o.socket,
    maxNodes, c.id, o.socktype)
  return c
}

func main() {
  nodeClient := Init(InitFlags())
  webClient := InitWebClient(nodeClient)

  go nodeClient.Start()
  go webClient.InitWebSocketServer("8080")
  webClient.InitHTTPServer("8000")
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
}

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
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
