#!/bin/bash

# Script used on remote elevator to set Gopath and run program
export GOPATH=$HOME/work
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
export GOROOT=
export GOBIN=$GOPATH/bin

clear

go run ~/work/src/github.com/andersliland/ttk4145-project/main.go
echo "run main.go"
