var Node = function() {
  this.name = "Monte Carlo";
  this.running = false;
}

Node.prototype.process = function(timeout) {
  console.log("Starting...");
  var endTime = new Date().getTime() + timeout;
  var inside = 0, total = 0;
  while ((new Date()).getTime() < endTime) {
    var x = Math.random();
    var y = Math.random();

    total += 1;
    if (((x * x) + (y * y)) < 1) {
      inside += 1; 
    }
  }
  console.log("Done.");
  return {inside: inside, total: total}
}

Node.prototype.montecarlo = function() {
  this.running = true;
  return this.process(10000);
}

Node.prototype.merge = function(results) {
  var inside = 0, total = 0;
  for (var i = 0; i < results.length; i++) {
    var result = results[i];
    inside += result.inside;
    total += result.total;
  }
  return (inside / total) * 4;
}

onmessage = function(msg) {
  console.log("Got message from spawner:", msg.data);
  var node = new Node();
  postMessage(node.montecarlo());
}
