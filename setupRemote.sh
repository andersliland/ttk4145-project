#!/bin/bash
clear
export GOPATH=$HOME/work
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
export GOROOT=
export GOBIN=$GOPATH/bin


go run ~/work/src/github.com/andersliland/ttk4145-project/main.go
echo "run main.go"
