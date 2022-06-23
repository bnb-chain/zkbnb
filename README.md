#zkbas

## goctl

```shell
# api
goctl api go -api xx.api -dir . -style gozero
# rpc
goctl rpc protoc xx.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

## Fast Deploy on Local Server
```shell
# make sure `docker`, `node`, `pm2`, `jq` installed and docker daemon started before the local deploy
# Refer to the comment section of `1-deploy-local.sh` for installation above

> cd /tmp && rm -rf ./zkbas ./zkbas.vk ./zkbas.pk && git clone --branch local_testnet_with_keccak256 https://github.com/bnb-chain/zkbas && sudo bash -x ./zkbas/deploy/local/1-deploy-local.sh
```

### Check status

`pm2 list`  check all monitors, rpc and prover service status
`pm2 logs blockMonitor` check logs of blockMonitor service, other services are similar, `pm2 logs committer`,  `pm2 logs l2BlockMonitor`,  `pm2 logs proverHub`  ...
check commiter tx on l1: `https://testnet.bscscan.com/address/0xC0E6DD1223b0446E2b794605d48234573472a1bc`
connect zkbas Database from office network :    
(refer: install psql on Mac: `brew install postgresql && sudo gem install pg` )

`psql -U postgres -d zecreylegend -W -h 172.22.41.148`
input password:  `ZecreyProtocolDB@123`

`\dt+` : check all tables on  zecreylegend Database

`select * from asset_info;`   check asset_info table, see if there are 3 assets in the table

`select * from account;`  check account  table, see if there are 4 accounts in the table

`select * from l2_tx_event_monitor;`  check l2_tx_event_monitor table, see if there are at least 8 tx events in the table. (4 RegisterZns events + 4 Deposit events   for treasury, gas, sher and gavin  )

`select * from l1_tx_sender;`   check l1_tx_sender table, see if there are at least 6 rows in the table. (three l1 CommitTxType(1)  txs + three l1 VerifyAndExecuteTxType(2)  txs,  three txs for 8 l2 blocks)

### L2 tx tests
For now, there are some tests in service/rpc/globalRPC/test dir, we can run as follows:

```shell
cd ~/zkbas-deploy/zkbas

go test ./service/rpc/globalRPC/test -run TestSendTransferTx -count=1

go test ./service/rpc/globalRPC/test -run TestSendCreateCollectionTx -count=1

go test ./service/rpc/globalRPC/test -run TestSendMintNftTx -count=1

go test ./service/rpc/globalRPC/test -run TestSendAtomicMatchTx -count=1

```
