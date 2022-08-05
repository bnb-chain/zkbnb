# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: zkbas android ios zkbas-cross evm all test clean
.PHONY: zkbas-linux zkbas-linux-386 zkbas-linux-amd64 zkbas-linux-mips64 zkbas-linux-mips64le
.PHONY: zkbas-linux-arm zkbas-linux-arm-5 zkbas-linux-arm-6 zkbas-linux-arm-7 zkbas-linux-arm64
.PHONY: zkbas-darwin zkbas-darwin-386 zkbas-darwin-amd64
.PHONY: zkbas-windows zkbas-windows-386 zkbas-windows-amd64

APP = ./service/api/app
EXPLORE = ./service/api/explorer

globalRPCProtoPath = ./service/rpc/globalRPC

app:
	cd $(APP) && goctl api go -api app.api -dir .;
	@echo "Done generate app api";


globalRPCProto:
	cd $(globalRPCProtoPath) && goctl rpc protoc globalRPC.proto --go_out=. --go-grpc_out=. --zrpc_out=.;
	@echo "Done generate globalRPCProto";


deploy:
	sudo bash -x ./deploy-local.sh new

test:
	sudo bash -x ./local-test.sh
