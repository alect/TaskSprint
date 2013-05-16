#!/usr/bin/python
from fabric.api import *
from fabric.operations import *
from fabric.exceptions import *
from cfg import *
import sys, time, boto.ec2, string, StringIO

class TimeStampStream(StringIO.StringIO):
  def write(self, s):
    if s.startswith("["):
      sys.stdout.write("[" + time.asctime() + "]")
    sys.stdout.write(s)

def start_coordinator(address, servers, index, program):
  print "\nStarting coordinator at", address
  command = "~/TaskSprint/bin/coordinator"
  command += " -servers " + servers
  command += " -me " + str(index)
  command += " -dc ~/TaskSprint/deploy/" + program
  with settings(host_string = address):
    run(command, stdout = TimeStampStream())

def start_client(address, servers, program):
  print "\nStarting client at", address
  command = "~/TaskSprint/bin/client"
  command += " -servers " + servers
  command += " -socket " + address + port
  command += " -network tcp"
  command += " -program ~/TaskSprint/deploy/" + program
  with settings(host_string = address):
    run(command, stdout = TimeStampStream())

def transfer_program(local_path, name, address):
  print "\nTransferring", name, "to", address
  remote_path = "~/TaskSprint/deploy/" + name
  with settings(host_string = address):
    while True:
      try:
        put(local_path, remote_path)
        run("chmod +x " + remote_path)
        break
      except NetworkError:
        time.sleep(2)

def refresh_and_build(address):
  print "\nUpdating TaskSprint build on", address
  with settings(host_string = address):
    with cd("~/TaskSprint"):
      while True:
        try:
          run("git pull", quiet = True)
          run("mkdir bin")
          run("go build -o bin/client client/src/client/client.go")
          run("go build -o bin/coordinator coordinator/src/coord/main/coord.go")
          run(link_library("python/client/TaskSprintNode.py"))
          run(link_library("python/coordinator/TaskSprintCoordinator.py"))
          run(link_library("python/jsonify.py"))
          break
        except NetworkError:
          time.sleep(2)

  print "Done!"

def link_library(path):
  split = string.split(path, "/")
  return "ln -s ~/TaskSprint/libraries/%s deploy/%s" %(path, split[len(split) - 1])

def setup(address, program_path, program_name):
  refresh_and_build(address)
  transfer_program(program_path, program_name, address)

def setup_coordinator(i, servers, program, deploy_dir):
  print "Initializing coordinator."
  addresses = string.split(servers, ",")
  address = addresses[i]
  program_path = get_program_path(program, deploy_dir)
  setup(address, program_path, program)

  addresses[i] = "0.0.0.0"
  servers = string.join(addresses, port + ",") + port
  start_coordinator(address, servers, i, program)

def setup_client(address, servers, program, deploy_dir):
  print "Initializing client."
  program_path = get_program_path(program, deploy_dir)
  setup(address, program_path, program)
  start_client(address, servers, program)

def get_program_path(program, deploy_dir):
  if program.startswith("/") or program.startswith("~"):
    return program
  return deploy_dir + "/" + program

if __name__ == "__main__":
  if len(sys.argv) != 6:
    print >> sys.stderr, "bad arguments"
    sys.exit(1)

  env.user = user
  env.key_filename = sys.argv[5] + "/" + key_name + ".pem"
  if sys.argv[1] == "cd":
    setup_coordinator(int(sys.argv[2]), sys.argv[3], sys.argv[4], sys.argv[5])
  else:
    setup_client(*sys.argv[2:])
