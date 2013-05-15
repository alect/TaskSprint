#!/usr/bin/python
from cfg import *
import sys, time, boto.ec2, string, subprocess, os

local_dir = os.getcwd()

def deploy(node, sched, clients, coords):
  print "Deploying %d %ss, %d %ss" %(clients, node, coords, sched)
  coord_reservation, client_reservation = init(coords), init(clients)

  coord_addresses = get_addresses(coord_reservation)
  servers = string.join(coord_addresses, ",")
  for i in xrange(len(coord_addresses)):
    start_coordinator(i, servers, sched)

  client_addresses = get_addresses(client_reservation)
  servers = string.join(coord_addresses, port + ",") + port
  for client_address in client_addresses:
    start_client(client_address, servers, node)

def as_terminal_run(command):
  ascript = 'tell application "Terminal" to do script "%s"' %command
  subprocess.call(["osascript", "-e", ascript])

def start_coordinator(i, servers, program):
  command = local_dir + "/start.py "
  command += "cd %s %s %s %s" %(str(i), servers, program, local_dir)
  as_terminal_run(command)

def start_client(address, servers, program):
  command = local_dir + "/start.py "
  command += "cl %s %s %s %s" %(address, servers, program, local_dir)
  as_terminal_run(command)

def init(number):
  conn = connect()
  reservation = start_instances(conn, number)
  print_reservation(reservation)
  return reservation

def get_addresses(reservation):
  all_dns = map(get_dns, reservation.instances)
  return all_dns

def get_dns(instance):
  print "\nWaiting for instance", instance.id, "to start..."
  instance.update()
  while instance.state_code == 0:
    time.sleep(2)
    instance.update()

  if instance.state_code != 16: # 16 = running
    print >> sys.stderr, "Instance", instance, "failed to start."

  print "Instance", instance.id, "is up!"
  return instance.public_dns_name

def print_reservation(reservation):
  print "\nReservation", reservation.id,  "by", reservation.owner_id
  print "----------------------------------------"
  print "Instances:", reservation.instances

def connect():
  return boto.ec2.connect_to_region("us-east-1",
      aws_access_key_id = access_key,
      aws_secret_access_key = secret_key)

def start_instances(conn, count):
  return conn.run_instances(ami,
      min_count = count,
      max_count = count,
      key_name = key_name,
      instance_type = instance_type,
      security_groups = security_groups)

if __name__ == "__main__":
  if len(sys.argv) != 5:
    usage = "usage: /path/to/node.py /path/to/scheduler.py"
    usage += " #clients #coordinators"
    example = "\nexample: node.py scheduler.py 10 4"
    print >> sys.stderr, usage, example
    sys.exit(1)
  deploy(sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]))
