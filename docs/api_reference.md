# API Reference

## Version: 1.0

### /

#### GET

##### Summary

Get status of zkbnb

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Status](#status) |

### /api/v1/account

#### GET

##### Summary

Get account by account's name, index or pk

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | name/index/pk | Yes | string |
| value | query | value of name/index/pk | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Account](#account) |

### /api/v1/accountPendingTxs

#### GET

##### Summary

Get pending transactions of a specific account

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | account_name/account_index/account_pk | Yes | string |
| value | query | value of account_name/account_index/account_pk | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Txs](#txs) |

### /api/v1/accountNfts

#### GET

##### Summary

Get nfts of a specific account

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | account_name/account_index/account_pk | Yes | string |
| value | query | value of account_name/account_index/account_pk | Yes | string |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Nfts](#nfts) |

### /api/v1/accountTxs

#### GET

##### Summary

Get transactions of a specific account

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | account_name/account_index/account_pk | Yes | string |
| value | query | value of account_name/account_index/account_pk | Yes | string |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Txs](#txs) |

### /api/v1/accounts

#### GET

##### Summary

Get accounts

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Accounts](#accounts) |

### /api/v1/asset

#### GET

##### Summary

Get asset

##### Parameters

| Name  | Located in | Description        | Required | Schema  |
|-------| ---------- |--------------------| -------- |---------|
| by    | query | id/symbol          | Yes | string  |
| value | query | value of id/symbol | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Asset](#assets) |

### /api/v1/assets

#### GET

##### Summary

Get assets

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Assets](#assets) |

### /api/v1/block

#### GET

##### Summary

Get block by its height or commitment

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | height/commitment | Yes | string |
| value | query | value of height/commitment | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Block](#block) |

### /api/v1/blockTxs

#### GET

##### Summary

Get transactions in a block

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | block_height/block_commitment | Yes | string |
| value | query | value of block_height/block_commitment | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Txs](#txs) |

### /api/v1/blocks

#### GET

##### Summary

Get blocks

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Blocks](#blocks) |

### /api/v1/currentHeight

#### GET

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [CurrentHeight](#currentheight) |

### /api/v1/gasAccount

#### GET

##### Summary

Get gas account, who will charge gas fees for transactions

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GasAccount](#gasaccount) |

### /api/v1/gasFee

#### GET

##### Summary

Get gas fee amount for using a specific asset as gas asset

##### Parameters

| Name     | Located in | Description | Required | Schema |
|----------| ---------- |-------------| -------- | ---- |
| asset_id | query | id of asset | Yes | integer |
| tx_type  | query | type of tx  | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GasFee](#gasfee) |

### /api/v1/gasFeeAssets

#### GET

##### Summary

Get supported gas fee assets

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GasFeeAssets](#gasfeeassets) |

### /api/v1/layer2BasicInfo

#### GET

##### Summary

Get zkbnb general info, including contract address, and count of transactions and active users

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Layer2BasicInfo](#layer2basicinfo) |

### /api/v1/maxOfferId

#### GET

##### Summary

Get max nft offer id for a specific account

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| account_index | query | index of account | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [MaxOfferId](#maxofferid) |

### /api/v1/pendingTxs

#### GET

##### Summary

Get pending transactions

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Txs](#txs) |

### /api/v1/executedTxs

#### GET

##### Summary

Get executed transactions

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |
| from_id | query | start from the id | No | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Txs](#txs) |


### /api/v1/nextNonce

#### GET

##### Summary

Get next nonce

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| account_index | query | index of account | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [NextNonce](#nextnonce) |

### /api/v1/search

#### GET

##### Summary

Search with a specific keyword

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| keyword | query | keyword | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Search](#search) |

### /api/v1/tx

#### GET

##### Summary

Get transaction by hash

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| hash | query | hash of tx | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [EnrichedTx](#enrichedtx) |

### /api/v1/txs

#### GET

##### Summary

Get transactions

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Txs](#txs) |

### /api/v1/sendTx

#### POST

##### Summary

Send raw transaction

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | raw tx | Yes | [ReqSendTx](#reqsendtx) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [TxHash](#txhash) |

### Models

#### Account

| Name               | Type                              | Description | Required |
|--------------------|-----------------------------------| ----------- | -------- |
| status             | integer                           |  | Yes |
| index              | long                              |  | Yes |
| name               | string                            |  | Yes |
| pk                 | string                            |  | Yes |
| nonce              | long                              |  | Yes |
| assets             | [ [AccountAsset](#accountasset) ] |  | Yes |
| total_asset_value  | string                            |  | Yes |

#### AccountAsset

| Name    | Type | Description | Required |
|---------| ---- | ----------- | -------- |
| id      | integer |  | Yes |
| name    | string |  | Yes |
| balance | string |  | Yes |
| price   | string |  | Yes |

#### Accounts

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | integer |  | Yes |
| accounts | [ [SimpleAccount](#simpleaccount) ] |  | Yes |

#### Asset

| Name         | Type    | Description | Required |
|--------------|---------| ----------- | -------- |
| id           | integer |  | Yes |
| name         | string  |  | Yes |
| decimals     | integer |  | Yes |
| symbol       | string  |  | Yes |
| address      | string  |  | Yes |
| price        | string  |  | Yes |
| is_gas_asset | integer |  | Yes |
| icon         | string  |  | Yes |

#### Assets

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | integer |  | Yes |
| assets | [ [Asset](#asset) ] |  | Yes |

#### Block

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| commitment | string |  | Yes |
| height | long |  | Yes |
| state_root | string |  | Yes |
| priority_operations | long |  | Yes |
| pending_on_chain_operations_hash | string |  | Yes |
| pending_on_chain_operations_pub_data | string |  | Yes |
| committed_tx_hash | string |  | Yes |
| committed_at | long |  | Yes |
| verified_tx_hash | string |  | Yes |
| verified_at | long |  | Yes |
| txs | [ [Tx](#tx) ] |  | Yes |
| status | long |  | Yes |
| size | long |  | Yes |

#### Blocks

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | integer |  | Yes |
| blocks | [ [Block](#block) ] |  | Yes |

#### ContractAddress

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | Yes |
| address | string |  | Yes |

#### CurrentHeight

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| height | long |  | Yes |

#### EnrichedTx

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
|  | [Tx](#tx) |  | No |
| committed_at | long |  | Yes |
| verified_at | long |  | Yes |
| executed_at | long |  | Yes |

#### GasAccount

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | long |  | Yes |
| index | long |  | Yes |
| name | string |  | Yes |

#### GasFee

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| gas_fee | string |  | Yes |

#### GasFeeAssets

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| assets | [ [Asset](#asset) ] |  | Yes |

#### Layer2BasicInfo

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| block_committed | long |  | Yes |
| block_verified | long |  | Yes |
| total_transaction_count | long |  | Yes |
| yesterday_transaction_count | long |  | Yes |
| today_transaction_count | long |  | Yes |
| yesterday_active_user_count | long |  | Yes |
| today_active_user_count | long |  | Yes |
| contract_addresses | [ [ContractAddress](#contractaddress) ] |  | Yes |

#### MaxOfferId

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| offer_id | long |  | Yes |

#### NextNonce

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| nonce | long |  | Yes |

#### Nft

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| index | long |  | Yes |
| creator_account_index | long |  | Yes |
| owner_account_index | long |  | Yes |
| content_hash | string |  | Yes |
| l1_address | string |  | Yes |
| l1_token_id | string |  | Yes |
| creator_treasury_rate | long |  | Yes |
| collection_id | long |  | Yes |

#### Nfts

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | long |  | Yes |
| nfts | [ [Nft](#nft) ] |  | Yes |

#### ReqGetAccount

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |

#### ReqGetAccountPendingTxs

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |

#### ReqGetAccountNfts

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |
| offset | [uint16](#uint16) |  | Yes |
| limit | [uint16](#uint16) |  | Yes |

#### ReqGetAccountTxs

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |
| offset | [uint16](#uint16) |  | Yes |
| limit | [uint16](#uint16) |  | Yes |

#### ReqGetAsset

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |

#### ReqGetBlock

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |

#### ReqGetBlockTxs

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |

#### ReqGetGasFee

| Name     | Type | Description | Required |
|----------| ---- | ----------- | -------- |
| asset_id | integer |  | Yes |
| tx_type       | integer |  | Yes |

#### ReqGetMaxOfferId

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| account_index | integer |  | Yes |

#### ReqGetNextNonce

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| account_index | integer |  | Yes |

#### ReqGetRange

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| offset | integer |  | Yes |
| limit | integer |  | Yes |

#### ReqGetTx

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| hash | string |  | Yes |

#### ReqSearch

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| keyword | string |  | Yes |

#### ReqSendTx

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| tx_type | integer |  | Yes |
| tx_info | string |  | Yes |

#### Search

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| data_type | integer | 2:account; 4:pk; 9:block; 10:tx | Yes |

#### SimpleAccount

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| index | long |  | Yes |
| name | string |  | Yes |
| pk | string |  | Yes |

#### Status

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | integer |  | Yes |
| network_id | integer |  | Yes |

#### Tx

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| hash | string |  | Yes |
| type | long |  | Yes |
| amount | string |  | Yes |
| info | string |  | Yes |
| status | long |  | Yes |
| index | long |  | Yes |
| gas_fee_asset_id | long |  | Yes |
| gas_fee | string |  | Yes |
| nft_index | long |  | Yes |
| asset_id | long |  | Yes |
| asset_name | string |  | Yes |
| native_adress | string |  | Yes |
| extra_info | string |  | Yes |
| memo | string |  | Yes |
| account_index | long |  | Yes |
| account_name | string |  | Yes |
| nonce | long |  | Yes |
| expire_at | long |  | Yes |
| block_height | long |  | Yes |
| created_at | long |  | Yes |
| state_root | string |  | Yes |

#### TxHash

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| tx_hash | string |  | Yes |

#### Txs

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | integer |  | Yes |
| txs | [ [Tx](#tx) ] |  | Yes |
