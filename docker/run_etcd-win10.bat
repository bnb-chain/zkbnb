@echo off
unset https_proxy http_proxy all_proxy && \
    rm -rf /tmp/etcd-data.tmp && mkdir -p /tmp/etcd-data.tmp && \
    docker run --rm \
    -p 2379:2379 \
    --mount type=bind,source=/tmp/etcd-data.tmp,destination=/etcd-data \
    --name etcd-gcr-v3.5.1 \
    -d quay.io/coreos/etcd:v3.5.1 \
    /usr/local/bin/etcd \
    --name s1 \
    --data-dir /etcd-data \
    --listen-client-urls http://0.0.0.0:2379 \
    --advertise-client-urls http://0.0.0.0:2379 \
    --listen-peer-urls http://0.0.0.0:2380 \
    --initial-advertise-peer-urls http://0.0.0.0:2380 \
    --initial-cluster s1=http://0.0.0.0:2380 \
    --initial-cluster-token tkn \
    --initial-cluster-state new \
    --log-level info \
    --logger zap \
    --log-outputs stderr