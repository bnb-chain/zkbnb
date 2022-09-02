# API Reference

## Version: 1.0

### /

#### GET
##### Summary

Get status of zkbas

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

### /api/v1/accountMempoolTxs

#### GET
##### Summary

Get mempool transactions of a specific account

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | account_name/account_index/account_pk | Yes | string |
| value | query | value of account_name/account_index/account_pk | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [MempoolTxs](#mempooltxs) |

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

### /api/v1/currencyPrice

#### GET
##### Summary

Get asset price by its symbol

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| by | query | symbol | Yes | string |
| value | query | value of symbol | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [CurrencyPrice](#currencyprice) |

### /api/v1/currencyPrices

#### GET
##### Summary

Get assets' prices

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [CurrencyPrices](#currencyprices) |

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

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| asset_id | query | id of asset | Yes | integer |

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

Get zkbas general info, including contract address, and count of transactions and active users

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Layer2BasicInfo](#layer2basicinfo) |

### /api/v1/lpValue

#### GET
##### Summary

Get liquidity pool amount for a specific liquidity pair

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| pair_index | query | index of pair | Yes | integer |
| lp_amount | query | lp amount | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [LpValue](#lpvalue) |

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

### /api/v1/mempoolTxs

#### GET
##### Summary

Get mempool transactions

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| offset | query | offset, min 0 and max 100000 | Yes | integer |
| limit | query | limit, min 1 and max 100 | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [MempoolTxs](#mempooltxs) |

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

### /api/v1/pair

#### GET
##### Summary

Get liquidity pool info by its index

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| index | query | index of pair | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Pair](#pair) |

### /api/v1/pairs

#### GET
##### Summary

Get liquidity pairs

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [Pairs](#pairs) |

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


### /api/v1/swapAmount

#### GET
##### Summary

Get swap amount for a specific liquidity pair and in asset amount

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| pair_index | query | index of pair | Yes | integer |
| asset_id | query | id of asset | Yes | integer |
| asset_amount | query | amount of asset | Yes | string |
| is_from | query | is from asset | Yes | boolean (boolean) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [SwapAmount](#swapamount) |

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

### /api/v1/withdrawGasFee

#### GET
##### Summary

Get withdraw gas fee amount for using a specific asset as gas asset

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| asset_id | query | id of asset | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [GasFee](#gasfee) |

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

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | integer |  | Yes |
| index | long |  | Yes |
| name | string |  | Yes |
| pk | string |  | Yes |
| nonce | long |  | Yes |
| assets | [ [AccountAsset](#accountasset) ] |  | Yes |

#### AccountAsset

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | integer |  | Yes |
| name | string |  | Yes |
| balance | string |  | Yes |
| lp_amount | string |  | Yes |

#### Accounts

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | integer |  | Yes |
| accounts | [ [SimpleAccount](#simpleaccount) ] |  | Yes |

#### Asset

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | integer |  | Yes |
| name | string |  | Yes |
| decimals | integer |  | Yes |
| symbol | string |  | Yes |
| address | string |  | Yes |
| is_gas_asset | integer |  | Yes |

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

#### CurrencyPrice

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| pair | string |  | Yes |
| asset_id | integer |  | Yes |
| price | string |  | Yes |

#### CurrencyPrices

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | integer |  | Yes |
| currency_prices | [ [CurrencyPrice](#currencyprice) ] |  | Yes |

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

#### LpValue

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| asset_a_id | integer |  | Yes |
| asset_a_name | string |  | Yes |
| asset_a_amount | string |  | Yes |
| asset_b_id | integer |  | Yes |
| asset_b_name | string |  | Yes |
| asset_b_amount | string |  | Yes |

#### MaxOfferId

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| offer_id | long |  | Yes |

#### MempoolTxs

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| total | integer |  | Yes |
| mempool_txs | [ [Tx](#tx) ] |  | Yes |

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

#### Pair

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| index | integer |  | Yes |
| asset_a_id | integer |  | Yes |
| asset_a_name | string |  | Yes |
| asset_a_amount | string |  | Yes |
| asset_b_id | integer |  | Yes |
| asset_b_name | string |  | Yes |
| asset_b_amount | string |  | Yes |
| fee_rate | long |  | Yes |
| treasury_rate | long |  | Yes |
| total_lp_amount | string |  | Yes |

#### Pairs

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| pairs | [ [Pair](#pair) ] |  | Yes |

#### ReqGetAccount

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |

#### ReqGetAccountMempoolTxs

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

#### ReqGetCurrencyPrice

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| by | string |  | Yes |
| value | string |  | Yes |

#### ReqGetGasFee

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| asset_id | integer |  | Yes |

#### ReqGetLpValue

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| pair_index | integer |  | Yes |
| lp_amount | string |  | Yes |

#### ReqGetMaxOfferId

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| account_index | integer |  | Yes |

#### ReqGetNextNonce

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| account_index | integer |  | Yes |

#### ReqGetPair

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| index | integer |  | Yes |

#### ReqGetRange

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| offset | integer |  | Yes |
| limit | integer |  | Yes |

#### ReqGetSwapAmount

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| pair_index | integer |  | Yes |
| asset_id | integer |  | Yes |
| asset_amount | string |  | Yes |
| is_from | boolean (boolean) |  | Yes |

#### ReqGetTx

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| hash | string |  | Yes |

#### ReqGetWithdrawGasFee

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| asset_id | integer |  | Yes |

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

#### SwapAmount

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| asset_id | integer |  | Yes |
| asset_name | string |  | Yes |
| asset_amount | string |  | Yes |

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
| pair_index | long |  | Yes |
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
