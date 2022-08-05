# zkbas


### goctl

```shell
# api
goctl api go -api xx.api -dir . -style gozero
# rpc
goctl rpc protoc xx.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

### mockgen

```shell
go install github.com/golang/mock/mockgen@v1.6.0
```