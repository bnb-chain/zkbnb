# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: zkbas android ios zkbas-cross evm all test clean
.PHONY: zkbas-linux zkbas-linux-386 zkbas-linux-amd64 zkbas-linux-mips64 zkbas-linux-mips64le
.PHONY: zkbas-linux-arm zkbas-linux-arm-5 zkbas-linux-arm-6 zkbas-linux-arm-7 zkbas-linux-arm64
.PHONY: zkbas-darwin zkbas-darwin-386 zkbas-darwin-amd64
.PHONY: zkbas-windows zkbas-windows-386 zkbas-windows-amd64
GOBIN?=${GOPATH}/bin

VERSION=$(shell git describe --tags)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_DATE=$(shell git log -n1 --pretty='format:%cd' --date=format:'%Y%m%d')
REPO=github.com/bnb-chain/zkbas
IMAGE_NAME=ghcr.io/bnb-chain/zkbas
API_SERVER = ./service/apiserver

api-server:
	cd $(API_SERVER) && ${GOBIN}/goctl api go -api server.api -dir .;
	@echo "Done generate server api";

deploy:
	sudo bash -x ./deploy-local.sh new

integration-test:
	sudo bash -x ./local-test.sh

test:
	@echo "--> Running go test"
	@go test ./...

tools:
	go install -u github.com/zeromicro/go-zero/tools/goctl@v1.4.0

build: api-server build-only

build-only:
	go build -o build/bin/zkbas -ldflags="-X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.gitDate=${GIT_COMMIT_DATE}" ./cmd/zkbas

docker-image:
	go mod vendor # temporary, should be removed after open source
	docker build . -t ${IMAGE_NAME}