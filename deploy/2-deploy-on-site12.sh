#!/bin/bash

# only for `root@tf_bsc_qa_bsc_zecrey_site12_ec2`
# config
echo '1. basic config and git clone repos'
yum install jq -y
npm install pm2 -g
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ~/zecreylegend-bak && mv ~/zecreylegend ~/zecreylegend-bak
mkdir zecreylegend && cd zecreylegend
git clone --branch qa_testnet_with_keccak256 https://github.com/bnb-chain/zecrey-legend

echo "2. run committer"
pm2 delete all

cd ~/zecreylegend/zecrey-legend
pm2 start --name committer "go run service/cronjob/committer/committer.go"
