#!/usr/bin/env bash

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin:/usr/local/bin
API_SERVER=./service/apiserver
cd $API_SERVER && goctl api go -api server.api -dir .