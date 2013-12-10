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
  cache TemplateCache
}

type IndexInfo struct {
  Name string
}

func InitClient() *WebClient {
  client := new(WebClient)
  client.cache = make(map[string]*template.Template)
  return client
}

func (client *WebClient) getTemplate(name string) *template.Template {
  t, ok := client.cache[name]
  if ok { return t }

  t, err := template.ParseFiles("web/" + name + ".html")
  if err != nil { log.Fatal("Error fetching template: ", err) }
  client.cache[name] = t
  return t
}

func (client *WebClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func initHTTPServer(port string) {
  client := InitClient()
  http.Handle("/", client)
  http.HandleFunc("/js/", handleStatic)
  http.HandleFunc("/css/", handleStatic)

  fmt.Println("HTTP listening on port " + port + "...")
  http.ListenAndServe(":" + port, nil)
}

func WebSocketHandler(ws *websocket.Conn) {
  websocket.Message.Send(ws, "Hey!")
  for {
    var msg string
    err := websocket.Message.Receive(ws, &msg)
    if err != nil { 
      fmt.Println("Error receiving."); 
      ws.Close();
      return; 
    }

    fmt.Println("Received:", msg)
    websocket.Message.Send(ws, "I heard you!")
  }
}

func initWebSocketServer(port string) {
  http.Handle("/live", websocket.Handler(WebSocketHandler))
  fmt.Println("Web Socket listening on port " + port + "...")
  http.ListenAndServe(":" + port, nil)
}

func main() {
  go initWebSocketServer("8080")
  initHTTPServer("8000")
}
