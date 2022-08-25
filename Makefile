# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: zkbas android ios zkbas-cross evm all test clean
.PHONY: zkbas-linux zkbas-linux-386 zkbas-linux-amd64 zkbas-linux-mips64 zkbas-linux-mips64le
.PHONY: zkbas-linux-arm zkbas-linux-arm-5 zkbas-linux-arm-6 zkbas-linux-arm-7 zkbas-linux-arm64
.PHONY: zkbas-darwin zkbas-darwin-386 zkbas-darwin-amd64
.PHONY: zkbas-windows zkbas-windows-386 zkbas-windows-amd64

API_SERVER = ./service/apiserver

api-server:
	cd $(API_SERVER) && goctl api go -api server.api -dir .;
	@echo "Done generate server api";

deploy:
	sudo bash -x ./deploy-local.sh new

integration-test:
	sudo bash -x ./local-test.sh

tools:
	go install github.com/zeromicro/go-zero/tools/goctl@v1.4.0

build: api-server
	go build -o build/api-server ./service/apiserver/server.go
	go build -o build/committer ./service/committer/main.go
	go build -o build/monitor ./service/monitor/main.go
	go build -o build/prover ./service/prover/main.go
	go build -o build/sender ./service/sender/main.go
	go build -o build/witness ./service/witness/main.go