#!/usr/bin/env bash

export GOPATH=`pwd`/../../../../

go run ${GOPATH}/src/github.com/agentinstaller/main/main.go -id=ID -key=KEY -master=http://xxxx:7777 -dir=xxx