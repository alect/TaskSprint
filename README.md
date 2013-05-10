TaskSprint (Name Not Finalized) 
==========

Introduction
--------------
TaskSprint is a general-purpose, fault-tolerant,
distributed system for deterministic and non-deterministic computation.
TaskSprint differentiates itself from existing systems, like MapReduce and Pig,
by allowing non-deterministic computations, allowing arbitrary stage
computations (as opposed to two in MapReduce), having no central point of
failure (as opposed to Pig/MapReduce's master), being language agnostic, and
being computation agnostic.

TaskSprint's core challenge is to provide all of these facilities in a
fault-tolerant manner while exposing minimal complexity to the user. To this
end, TaskSprint's only requirement from the user are two small pieces of
software: a scheduler and a node. The scheduler's job is to define a schedule
for tasks, and the node's job is implement those tasks. All the intermediate
  work is handled by the system. This includes distributing and replicating the
  computation across nodes in the network, handling data shuffling efficiently,
  handling task dependencies, and dealing with scheduler, node, and network
  failures.

Design
----------------
TaskSprint's infrastructure is centered on two major components: the coordinator
and the client. The coordinator communicates with the user's scheduler, and the
client communicate's with the user's nodes. The coordinator and the client
communicate with each other to execute the user's computation. To support
replicated, non-deterministic computations, we seed PRNGs with the same value.

Coordinator
--------------
The coordinator is a Paxos-replicated, event-driven state-machine that responds
to clients' requests for the current task schedule. Clients query the
coordinator for the latest "view", the task schedule, and notify the
coordinator when tasks are completed by nodes. A view includes a mapping of
tasks to clients in addition to task parameters as defined by the user's
scheduler. Parameters include task name, task arguments, and task dependencies. 

The coordinator monitors the liveliness of a clients. When a client hasn't
queried the coordinator for a certain period of time, the coordinator marks the
client as dead and reassigns its tasks. A single replica known as the ``leader''
periodically inserts "TICK" operations into the Paxos log to ensure that all
replicas exactly agree on when clients died in relation to other operations. New
leaders are periodically elected via Paxos to handle leader failure.

The coordinator communicates events to the user's scheduler. In response to
events, the scheduler can start new tasks, kill existing tasks, or end the
computation. These event handlers are very simple and easy to implement while
providing developers a large amount of flexibility for their computations. For
instance, a genetic algorithm scheduler is able to group similar solutions for
crossover and a Bitcoin mining scheduler is able to kill tasks to free resources
when solutions are found.

Client
--------------
The client's main tasks are to launch nodes, schedule tasks on nodes, and handle
task dependencies, which may involve efficiently transfering data between nodes
and clients. To launch nodes, the client detects the amount of available
resources on the machine it is running on and launches a number of nodes
proportional to the amount of available resources. The client-node communication
is done via IPC using a language-agonistic, JSON driven protocol.

A client is alerted of new tasks by polling the coordinator for the most recent
view. The client checks for changes to task assignments, finds tasks intended
for it, and identifies any dependencies for its tasks. If there are
  dependencies, a client must satisfy them before executing the task. To do
  this, the dependant client communicates with the supplying clients responsible
  for the pre-requisite tasks. If nodes on a supplying client have completed the
    pre-requisite task, the needed data, if any, is transferred to the dependant
    client. This is done on a rolling basis such that data is transferred
    asynchronously as soon as a pre-requisite tasks are complete.

Once the client has obtained the necessary input data for a task, either by
satisfying dependencies or directly from the scheduler, the data is serialized
according to the JSON driven, language-agnostic protocol and the task is
executed on an available node. Upon completion, a node notifies its client with
the results. The client stores the results so that is may respond satisfy future
dependencies and notifies the coordinator about the task's completion.

Examples
--------------

### Monte Carlo

#### Node Code:
    #!/usr/bin/python
    from TaskSprintNode import *
    import random

    class MonteCarloNode(TaskSprintNode):
      @timeout(5)
      def process(self):
        while True:
          x, y = random.random(), random.random()
          if x**2 + y**2 < 1: self.inside += 1
          self.total += 1

      def montecarlo(self):
        self.inside, self.total = 0, 0
        try:
          self.process()
        except TimeoutError:
          pass
        return {"result" : {"inside" : self.inside, "total" : self.total}}

      def merge(self, *results):
        inside, total = 0, 0
        for result in results:
          inside += result["inside"]
          total += result["total"]
        return {"result" : (inside / float(total)) * 4}

    MonteCarloNode().start()

#### Scheduler Code:
    #!/usr/bin/python
    from TaskSprintCoordinator import *

    class MonteCarloCoordinator(TaskSprintCoordinator):
      def init(self, seed):
        alltasks = [self.start_task(name = "montecarlo") for i in xrange(11)]

        self.final = self.start_task(
          name = "merge",
          prekeys = ["result" for i in range(len(alltasks))],
          pretasks = alltasks,
          keys = ["result"]
        )

      def client_joined(self, cid):
        print "Somebody joined! %d" %cid

      def client_dead(self, cid):
        print "Somebody died :( %d" %cid

      def task_done(self, tid, values):
        if tid == self.final:
          self.finish(taskids = [tid], values = values)
        print "Task %d is done. Result: %s" %(tid, values)

    MonteCarloCoordinator().start()

Wait, what is it again?
------------------
A distributed computation system!
