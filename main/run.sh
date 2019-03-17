#!/usr/bin/env bash

export GOPATH=`pwd`/../../../../..

go run ${GOPATH}/src/github.com/TeaWeb/agentinstaller/main/main.go -id=ID -key=KEY -master=http://xxx:7777 -dir=xxx