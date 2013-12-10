var socket = new WebSocket("ws://localhost:8080/live");

function sendHi() {
  if (socket.readyState == 1) {
    socket.send("hello!");
    setTimeout(sendHi, 1000);
  }
}

socket.onopen = function() {
  console.log("Socket is open.");
  var worker = new Worker("js/node.js");
  worker.onmessage = function(msg) {
    console.log("Got message from worker:", msg.data);
    socket.send(msg.data);
  }
  worker.postMessage("work!");
}

socket.onmessage = function(msg) {
  console.log("Received:", msg.data);
}

socket.onclose = function() {
  console.log("Socket closed.");
}
