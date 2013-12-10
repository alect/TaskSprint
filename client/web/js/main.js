var socket = null;
var worker = null;

function workerMessage(msg) {
  console.log("Got message from worker:", msg.data);
  var stringData = JSON.stringify(msg.data);
  updateStatus("Finished! Result: " + stringData);
  if (socket.readyState == 1) {
    socket.send(stringData);
  }
}

function startWorker() {
  worker = new Worker("js/node.js");
  worker.onmessage = workerMessage;
}

function killWorker() {
  console.log("Killing worker.");
  worker.terminate();
}

function restartWorker() {
  if (worker != null) killWorker();
  startWorker();
}

function initSocket() {
  updateStatus("Connecting...");
  socket = new WebSocket("ws://localhost:8080/live");

  socket.onopen = function() {
    console.log("Socket is open.");
    updateStatus("Connected.");
    startWorker();
  }

  socket.onmessage = function(msg) {
    try {
      var task = JSON.parse(msg.data);
      if (task[0] !== "kill") {
        console.log("Received Task:", task);
        updateStatus("Starting task: " + task[0]);
        worker.postMessage(task);
        updateStatus("Started. Working...");
      } else {
        restartWorker();
      }
    } catch(err) {
      console.error("Error processing server message.");
    }
  }

  socket.onclose = function() {
    console.log("Socket closed.");
  }
}

function updateStatus(message) {
  statusDiv = document.getElementById("status");

  bElement = document.createElement("b");
  bElement.textContent = "Status (" + new Date().toUTCString() + "): ";
  statusDiv.appendChild(bElement);

  sElement = document.createElement("span");
  sElement.textContent = message;
  statusDiv.appendChild(sElement);
  statusDiv.appendChild(document.createElement("br"));
}

var computeButton = document.getElementById("compute");
computeButton.addEventListener("click", function() {
  computeButton.disabled = true;
  initSocket();
});
