package coordinator 

import "log"
import "os/exec"
import "io"
import "os"
import "strconv"
import "math/big"
import "crypto/rand"
import "encoding/json"
import "fmt"
import "net"
import "time"
import "strings"

type AllCoordinator struct {
  program string
  devSocket string
  socket string
  id int64
  l net.Listener
  co *Coordinator

  result int // ONLY FOR TESTING
}

type StartTaskParams struct {
  Name string
  Base []interface{}
  Keys []string
}

type FinishParams struct {
  TaskIDs []TaskID
  Values map[string]interface{} // only for testing!!
}

func (ac *AllCoordinator) startServer() {
  os.Remove(ac.socket) // For now...
  l, e := net.Listen("unix", ac.socket);
  if e != nil {
    log.Fatal("listen error: ", e);
  }
  ac.l = l

  for {
    conn, err := ac.l.Accept()
    if err != nil {
      fmt.Printf("RPC Error: %v\n", err.Error())
      ac.l.Close()
    } else {
      go ac.handleConnection(conn)
    }
  }
}

func (ac *AllCoordinator) handleConnection(conn net.Conn) {
  // Waiting for result
  buffer := make([]byte, 1024)
  size, readerr := conn.Read(buffer)
  if readerr != nil { log.Fatal(readerr) }

  // Unserializing and marking as finished
  queryString := string(buffer[:size])
  sepIndex := strings.Index(queryString, ":")
  trigger := queryString[:sepIndex]
  json := queryString[sepIndex + 1:]
  ac.handleQuery(trigger, json, conn)
}

func (ac *AllCoordinator) handleQuery(trigger, jsons string, conn net.Conn) {
  fmt.Printf("Received query for %s with params %s\n", trigger, jsons)
  if trigger == "start_task" {
    args := &StartTaskParams{}
    parseerr := json.Unmarshal([]byte(jsons), args)
    if parseerr != nil { log.Fatal(parseerr) }
    tid := ac.startTask(args)
    fmt.Fprintf(conn, "{\"tid\" : %d }", tid)
  } else if trigger == "finish" {
    args := &FinishParams{}
    parseerr := json.Unmarshal([]byte(jsons), args)
    if parseerr != nil { log.Fatal(parseerr) }
    ac.finish(args)
  }
  conn.Close()
}

func (ac *AllCoordinator) startTask(args *StartTaskParams) TaskID {
  params := TaskParams{}
  params.FuncName = args.Name
  params.DoneKeys = args.Keys
  /* params.PreReqTasks = prereqTasks */
  /* params.PreReqKey = prereqKeys */
  params.BaseObject = args.Base
  task := ac.co.StartTask(params)
  return task
}

func (ac *AllCoordinator) finish(args *FinishParams) {
  ac.co.Finish(args.TaskIDs)

  ac.result = int(args.Values["result"].(float64))
}

func (ac *AllCoordinator) Result() int {
  return ac.result
}

func (ac *AllCoordinator) startProgram() {
  // Starting the subproc and copying its stdout to mine
  cmd := exec.Command(ac.program, ac.devSocket, ac.socket)
  stdout, outerr := cmd.StdoutPipe()
  if outerr != nil { log.Fatal(outerr) }
  stderr, errerr := cmd.StderrPipe()
  if errerr != nil { log.Fatal(errerr) }
  if err := cmd.Start(); err != nil { log.Fatal(err) }
  go io.Copy(os.Stdout, stdout)
  go io.Copy(os.Stderr, stderr)
}

func (ac *AllCoordinator) generateSocket(name string) string {
  s := "/tmp/tscoordinator-"
  s += strconv.FormatInt(int64(ac.id), 10) + "/"
  os.Mkdir(s, 0777)
  s +=  name
  return s
}

func (ac *AllCoordinator) clearNodeSockets() {
  s := "/tmp/tsclient-"
  s += strconv.FormatInt(int64(ac.id), 10) + "/"
  os.RemoveAll(s)
}

func getJson(task string, data interface{}) string {
  jsonString := ""
  if data != nil {
    byteArray, err := json.Marshal(data)
    if err != nil { log.Fatal("Failed to marshal data.") }
    jsonString = string(byteArray)
  }

  return fmt.Sprintf("[\"%s\", %s]", task, jsonString)
}

func (ac *AllCoordinator) trigger(task string, data interface{}) {
  fmt.Printf("Triggering: %s with %v\n", task, data)
  dataString := getJson(task, data)

  // Trying to connect to python
  conn, err := net.Dial("unix", ac.devSocket)
  for tries := 0; tries < 20 && err != nil; tries++ {
    time.Sleep(250 * time.Millisecond)
    conn, err = net.Dial("unix", ac.devSocket)
  }

  // Unsuccessful; die
  if err != nil { log.Fatal("Python is dead.", err) }

  // Sending data 
  fmt.Fprintf(conn, dataString)
  conn.Close()

  // Waiting for result
  /* buffer := make([]byte, 1024) */
  /* size, readerr := conn.Read(buffer) */
  /* if readerr != nil { log.Fatal(readerr) } */

  // Unserializing and marking as finished
  /* result := make(map[string]interface{}) */
  /* parseerr := json.Unmarshal(buffer[:size], &result) */
  /* if parseerr != nil { log.Fatal(parseerr) } */
}

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := rand.Int(rand.Reader, max)
  x := bigx.Int64()
  return x
}

func (ac *AllCoordinator) Init(co *Coordinator, seed int64) {
  ac.trigger("init", []int64{seed})
  ac.co = co;
}

func (ac *AllCoordinator) TaskDone(co *Coordinator,
tid TaskID, values map[string]interface{}) {
  ac.trigger("task_done", []interface{}{int64(tid), values})
}

func (ac *AllCoordinator) ClientJoined(co *Coordinator, cid ClientID) {
  ac.trigger("client_joined", []int64{int64(cid)})
}

func (ac *AllCoordinator) ClientDead(co *Coordinator, cid ClientID) {
  ac.trigger("client_dead", []int64{int64(cid)})
}

func MakeAllCoordinator(program string) *AllCoordinator {
  ac := &AllCoordinator{program, "", "", nrand(), nil, nil, 0}
  ac.devSocket = ac.generateSocket("devsocket")
  ac.socket = ac.generateSocket("coordsocket")
  go ac.startServer()
  ac.startProgram()
	return ac
}
