#!/bin/bash


# ./build.sh <TAG_VERSION> <REPO_PATH>

api="app explorer"
rpc="globalRPC proverHub"
cronjob="blockMonitor mempoolMonitor committer sender l2BlockMonitor proverClient governanceMonitor"

# pull newest code
cd $2
git fetch -unf wenki $1:refs/tags/$1
git checkout $1

# run goctl
for val in $api; do
    cd ./service/api/${val}
    echo "[${val}]: "
    goctl api go -api ${val}.api -dir . -style gozero
    cd ../../..
done

for val in $rpc; do
    cd ./service/rpc/${val}
    echo -n "[${val}]: "
    goctl rpc protoc ${val}.proto --go_out=. --go-grpc_out=. --zrpc_out=.
    cd ../../..
done

# go build all service&rpc in one script
for val in $api; do
    echo "Go Build [${val}]: "
    declare -l lower="${val}"
    go build -ldflags '-linkmode "external" -extldflags "-static"' -o ./bin/${lower} service/api/${val}/${lower}.go
done

for val in $rpc; do
    echo "Go Build [${val}]: "
    declare -l lower="${val}"
    go build -ldflags '-linkmode "external" -extldflags "-static"' -o ./bin/${lower} service/rpc/${val}/${lower}.go
done

for val in $cronjob; do
    echo "Go Build [${val}]: "
    declare -l lower="${val}"
    go build -ldflags '-linkmode "external" -extldflags "-static"' -o ./bin/${lower} service/cronjob/${val}/${lower}.go
done

gcloud auth configure-docker us-central1-docker.pkg.dev

for val in $api; do
    echo "Docker Build & Push [${val}]: "
    declare -l lower="${val}"
    docker build -t us-central1-docker.pkg.dev/zecrey-330903/zecrey-webhook/${lower}:$1 -f service/api/${val}/Dockerfile .
    docker push us-central1-docker.pkg.dev/zecrey-330903/zecrey-webhook/${lower}:$1
    docker image prune --filter label=stage=gobuilder --force
done

for val in $rpc; do
    echo "Docker Build & Push [${val}]: "
    declare -l lower="${val}"
    docker build -t us-central1-docker.pkg.dev/zecrey-330903/zecrey-webhook/${lower}:$1 -f service/rpc/${val}/Dockerfile .
    docker push us-central1-docker.pkg.dev/zecrey-330903/zecrey-webhook/${lower}:$1
    docker image prune --filter label=stage=gobuilder --force
done

for val in $cronjob; do
    echo "Docker Build & Push [${val}]: "
    declare -l lower="${val}"
    docker build -t us-central1-docker.pkg.dev/zecrey-330903/zecrey-webhook/${lower}:$1 -f service/cronjob/${val}/Dockerfile .
    docker push us-central1-docker.pkg.dev/zecrey-330903/zecrey-webhook/${lower}:$1
    docker image prune --filter label=stage=gobuilder --force
done

gcloud container clusters get-credentials "webhook" --region=us-central1-c
export TAG_NAME=$1
envsubst < ./kubeyaml/compiled.yaml | kubectl apply -f -