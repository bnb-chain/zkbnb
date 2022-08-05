# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: zecrey android ios zecrey-cross evm all test clean
.PHONY: zecrey-linux zecrey-linux-386 zecrey-linux-amd64 zecrey-linux-mips64 zecrey-linux-mips64le
.PHONY: zecrey-linux-arm zecrey-linux-arm-5 zecrey-linux-arm-6 zecrey-linux-arm-7 zecrey-linux-arm64
.PHONY: zecrey-darwin zecrey-darwin-386 zecrey-darwin-amd64
.PHONY: zecrey-windows zecrey-windows-386 zecrey-windows-amd64

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
