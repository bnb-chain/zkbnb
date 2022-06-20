#!/bin/bash

pm2 delete all

# only for `root@tf_bsc_qa_bsc_zecrey_site13_ec2`
# config
echo '1. basic config and git clone repos'
yum install wget -y
npm install pm2 -g
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ~/zecreylegend-bak && mv ~/zecreylegend ~/zecreylegend-bak
mkdir zecreylegend && cd zecreylegend
git clone --branch qa_testnet_with_keccak256 https://github.com/bnb-chain/zecrey-legend


echo "2. download zecrey-legend.vk and zecrey-legend.pk"
echo "You need to manually obtain the download link upload_vk_url from root@tf_bsc_qa_bsc_zecrey_site12_ec2:/root/zecreylegend/upload_vk_url.txt"
# curl $upload_vk_url -o ~/zecreylegend/zecrey-legend/zecrey-legend.vk

echo "You need to manually obtain the download link upload_pk_url from root@tf_bsc_qa_bsc_zecrey_site12_ec2:/root/zecreylegend/upload_pk_url.txt"
# curl $upload_pk_url -o ~/zecreylegend/zecrey-legend/zecrey-legend.pk

echo "3. run proverHub"
cd /zecrey-legend
pm2 start --name proverHub "go run service/rpc/proverHub/proverhub.go"

echo "4. run proverClient"
cd ~/zecreylegend/zecrey-legend
pm2 start --name proverClient "go run service/cronjob/proverClient/proverclient.go"

echo "5. run sender"
cd ~/zecreylegend/zecrey-legend
pm2 start --name sender "go run service/cronjob/sender/sender.go"
