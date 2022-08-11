# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: zkbas android ios zkbas-cross evm all test clean
.PHONY: zkbas-linux zkbas-linux-386 zkbas-linux-amd64 zkbas-linux-mips64 zkbas-linux-mips64le
.PHONY: zkbas-linux-arm zkbas-linux-arm-5 zkbas-linux-arm-6 zkbas-linux-arm-7 zkbas-linux-arm64
.PHONY: zkbas-darwin zkbas-darwin-386 zkbas-darwin-amd64
.PHONY: zkbas-windows zkbas-windows-386 zkbas-windows-amd64

APP = ./service/api/app

app:
	cd $(APP) && goctl api go -api app.api -dir .;
	@echo "Done generate app api";

deploy:
	sudo bash -x ./deploy-local.sh new

test:
	sudo bash -x ./local-test.sh


all:
	go build  -o build/app ./service/api/app/app.go
	go build -o build/committer ./service/cronjob/committer/committer.go
	go build -o build/monitor ./service/cronjob/monitor/monitor.go
	go build -o build/prover ./service/cronjob/prover/prover.go
	go build -o build/sender ./service/cronjob/sender/main.go
	go build -o build/witnessGenerator ./service/cronjob/witness/main.go