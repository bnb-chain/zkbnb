# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: zkbnb android ios zkbnb-cross evm all test clean
.PHONY: zkbnb-linux zkbnb-linux-386 zkbnb-linux-amd64 zkbnb-linux-mips64 zkbnb-linux-mips64le
.PHONY: zkbnb-linux-arm zkbnb-linux-arm-5 zkbnb-linux-arm-6 zkbnb-linux-arm-7 zkbnb-linux-arm64
.PHONY: zkbnb-darwin zkbnb-darwin-386 zkbnb-darwin-amd64
.PHONY: zkbnb-windows zkbnb-windows-386 zkbnb-windows-amd64
GOBIN?=${GOPATH}/bin

VERSION=$(shell git describe --tags)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_DATE=$(shell git log -n1 --pretty='format:%cd' --date=format:'%Y%m%d')
REPO=github.com/bnb-chain/zkbnb
IMAGE_NAME=ghcr.io/bnb-chain/zkbnb
API_SERVER = ./service/apiserver

api-server:
	cd $(API_SERVER) && goctl api go -api server.api -dir .;
	@echo "Done generate server api";

deploy:
	sudo bash -x ./deploy-local.sh new

integration-test:
	sudo bash -x ./local-test.sh

test: api-server
	@echo "--> Running go test"
	@go test ./...

tools:
	go install -u github.com/zeromicro/go-zero/tools/goctl@v1.4.0

build: api-server build-only

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0
	golangci-lint run ./...

build-only:
	go build -o build/bin/zkbnb -ldflags="-X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.gitDate=${GIT_COMMIT_DATE}" ./cmd/zkbnb

docker-image:
	go mod vendor # temporary, should be removed after open source
	docker build . -t ${IMAGE_NAME}

.PHONY: api-server deploy integration-test test tools build lint build-only docker-image
