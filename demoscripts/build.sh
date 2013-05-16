#!/bin/sh
export GOPATH=`pwd`/coordinator:`pwd`/client:$GOPATH
go install client
go install coord/main
