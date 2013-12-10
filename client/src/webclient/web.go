package main

import (
  "fmt"
  "log"
  "net/http"
  "html/template"
  "io/ioutil"
  "regexp"
  "code.google.com/p/go.net/websocket"
  /* "time" */
)

var validStatic = regexp.MustCompile("^/(js|css)/([a-zA-Z0-9]+)\\.(css|js)$")

type TemplateCache map[string]*template.Template

type WebClient struct {
  nodeClient *Client
  httpHandler *WebHTTPHandler
  wsHandler *WebWSHandler
}

type WebHTTPHandler struct {
  cache TemplateCache
}

type WebWSHandler struct {
  webClient *WebClient
  handler *websocket.Handler
  sockets map[int]*websocket.Conn
}

type IndexInfo struct {
  Name string
}

func InitClient() *WebClient {
  client := new(WebClient)
  client.httpHandler = new(WebHTTPHandler)
  client.httpHandler.cache = make(map[string]*template.Template)
  client.wsHandler = InitWebWSHandler(client)
  return client
}

func (client *WebHTTPHandler) getTemplate(name string) *template.Template {
  t, ok := client.cache[name]
  if ok { return t }

  t, err := template.ParseFiles("web/" + name + ".html")
  if err != nil { log.Fatal("Error fetching template: ", err) }
  client.cache[name] = t
  return t
}

func (client *WebHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  t := client.getTemplate("index")
  t.Execute(w, &IndexInfo{ "Sergio" })
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
  m := validStatic.FindString(r.URL.Path)
  if m == "" { http.NotFound(w, r); return }

  content, err := ioutil.ReadFile("web/" + m)
  if err != nil { log.Fatal("Error fetching static file: ", err) }
  fmt.Fprintf(w, string(content))
}

func InitWebClient(nodeClient *Client) *WebClient {
  client := InitClient()
  client.nodeClient = nodeClient;
  return client
}

func (client *WebClient) InitHTTPServer(port string) {
  http.HandleFunc("/js/", handleStatic)
  http.HandleFunc("/css/", handleStatic)
  http.Handle("/", client.httpHandler)

  fmt.Println("HTTP listening on port " + port + "...")
  http.ListenAndServe(":" + port, nil)
}

func (handler *WebWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  fmt.Println("Trying to serve web socket traffic...")
  handler.handler.ServeHTTP(w, r)
}

func InitWebWSHandler(webClient *WebClient) *WebWSHandler {
  wsHandler := new(WebWSHandler)
  wsHandler.sockets = make(map[int]*websocket.Conn)
  wsHandler.webClient = webClient;

  handler := websocket.Handler(func(ws *websocket.Conn) {
    fmt.Println("Hello, from the handler!")
    nodeClient := wsHandler.webClient.nodeClient

    nodeClient.viewMu.Lock()
    node := &Node{ ws, Free, nil, make(chan *Task), make(chan string) }
    nodeClient.nodes = append(nodeClient.nodes, node)
    nodeClient.viewMu.Unlock()

    for task := range node.taskChannel {
      // Received task from channel, dispatch it
      data := nodeClient.fetchParams(&task.params)
      websocket.Message.Send(ws, data)

      var msg string
      err := websocket.Message.Receive(ws, &msg)
      if err != nil {
        fmt.Println("Error receiving.");
        node.finishChannel <- "quit"
        ws.Close();
        return;
      }

      node.finishChannel <- msg;
    }
  });
  wsHandler.handler = &handler;
  return wsHandler;
}

func (client *WebClient) InitWebSocketServer(port string) {
  http.Handle("/live", client.wsHandler)
  fmt.Println("Web Socket listening on port " + port + "...")
  http.ListenAndServe(":" + port, nil)
}
