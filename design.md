# Zecrey-legend Design

## Tree(height)

- Account tree(32)

  - Node info: 

    - ```go
      type AccountNode struct{
          AccountIndex int64
          AccountNameHash string // bytes32
          PubKey string // bytes32
          Nonce int64
          AssetRoot string // bytes32
      }
      ```

  - Asset tree:

    - Node info:

      - ```go
        type AssetNode struct{
            AssetId int64
            Balance string
            LpAmount string
        }
        ```

- Liquidity tree(16)

  - Node info:

    - ```go
      type LiquidityNode struct{
          PairIndex int64
          AssetAId int64
          AssetABalance string
          AssetBId int64
          AssetBBalance string
      }
      ```

- Nft tree(40)

  - Node info:

    - ```go
      type NftNode struct{
          NftIndex int64
          NftContentHash string // bytes32
          CreatorAccountIndex int64
          OwnerAccountIndex int64
          AssetId int64
          AssetAmount string
          CreatorTreasuryRate int64 // 1% = 0.01 * 10000
          NftL1Address string
          NftL1TokenId string // uint256
      }
      ```

## Txs

### EmptyTx

EmptyTx is used for padding transactions in a block.

pubdata:

| Name   | Size(byte) | Comment          |
| ------ | ---------- | ---------------- |
| TxType | 1          | transaction type |

### RegisterZNS

This is a layer-1 transaction and a user needs to call this method first to register a layer-2 account.

pubdata:

| Name            | Size(byte) | Comment                        |
| --------------- | ---------- | ------------------------------ |
| TxType          | 1          | transaction type               |
| AccountName     | 32         | account name                   |
| AccountNameHash | 32         | hash value of the account name |

### CreatePair

This is a layer-1 transaction and is used for creating pair index for the layer-2 liquidity tree.

pubdata:

| Name      | Size(byte) | Comment            |
| --------- | ---------- | ------------------ |
| TxType    | 1          | transaction type   |
| PairIndex | 2          | trading pair index |
| AssetAId  | 2          | pair asset A index |
| AssetBId  | 2          | pair asset B index |

### Deposit

This is a layer-1 transaction and is used for depositing assets into the layer-2 account.

pubdata:

| Name            | Size(byte) | Comment           |
| --------------- | ---------- | ----------------- |
| TxType          | 1          | transaction type  |
| AccountIndex    | 4          | account index     |
| AccountNameHash | 32         | account name hash |
| AssetId         | 2          | asset index       |
| AssetAmount     | 16         | state amount      |

### DepositNft

This is a layer-1 transaction and is used for depositing nfts into the layer-2 account.

pubdata:

| Name            | Size(byte) | Comment               |
| --------------- | ---------- | --------------------- |
| TxType          | 1          | transaction type      |
| AccountIndex    | 4          | account index         |
| AccountNameHash | 32         | accout name hash      |
| NftIndex        | 5          | unique index of a nft |
| NftContentHash  | 32         | nft content hash      |
| NftL1Address    | 20         | nft layer-1 address   |
| NftL1TokenId    | 32         | nft layer-1 token id  |

### Transfer

This is a layer-2 transaction and is used for transfering assets in the layer-2 network.

pubdata:

| Name               | Size(byte) | Comment                |
| ------------------ | ---------- | ---------------------- |
| TxType             | 1          | transaction type       |
| FromAccountIndex   | 4          | from account index     |
| ToAccountIndex     | 4          | receiver account index |
| AssetId            | 2          | asset index            |
| AssetAmount        | 5          | packed asset amount    |
| GasFeeAccountIndex | 4          | gas fee account index  |
| GasFeeAssetId      | 2          | gas fee asset id       |
| GasFeeAssetAmount  | 2          | packed fee amount      |
| CallDataHash       | 32         | call data hash         |

### Swap

This is a layer-2 transaction and is used for making a swap for assets in the layer-2 network.

pubdata:

| Name                 | Size(byte) | Comment                |
| -------------------- | ---------- | ---------------------- |
| TxType               | 1          | transaction type       |
| FromAccountIndex     | 4          | from account index     |
| PairIndex            | 2          | unique pair index      |
| AssetAAmount         | 5          | packed asset amount    |
| AssetBAmount         | 5          | packed asset amount    |
| TreasuryAccountIndex | 4          | treasury account index |
| TreasuryFeeAmount    | 2          | packed fee amount      |
| GasFeeAccountIndex   | 4          | gas fee account index  |
| GasFeeAssetId        | 2          | gas fee asset id       |
| GasFeeAssetAmount    | 2          | packed fee amount      |

### AddLiquidity

This is a layer-2 transaction and is used for adding liquidity for a trading pair in the layer-2 network.

pubdata:

| Name               | Size(byte) | Comment               |
| ------------------ | ---------- | --------------------- |
| TxType             | 1          | transaction type      |
| FromAccountIndex   | 4          | from account index    |
| PairIndex          | 2          | unique pair index     |
| AssetAAmount       | 5          | packed asset amount   |
| AssetBAmount       | 5          | packed asset amount   |
| LpAmount           | 5          | packed asset amount   |
| GasFeeAccountIndex | 4          | gas fee account index |
| GasFeeAssetId      | 2          | gas fee asset id      |
| GasFeeAssetAmount  | 2          | packed fee amount     |

### RemoveLiquidity

This is a layer-2 transaction and is used for removing liquidity for a trading pair in the layer-2 network.

pubdata:

| Name               | Size(byte) | Comment               |
| ------------------ | ---------- | --------------------- |
| TxType             | 1          | transaction type      |
| FromAccountIndex   | 4          | from account index    |
| PairIndex          | 2          | unique pair index     |
| AssetAAmount       | 5          | packed asset amount   |
| AssetBAmount       | 5          | packed asset amount   |
| LpAmount           | 5          | packed asset amount   |
| GasFeeAccountIndex | 4          | gas fee account index |
| GasFeeAssetId      | 2          | gas fee asset id      |
| GasFeeAssetAmount  | 2          | packed fee amount     |

### Withdraw

This is a layer-2 transaction and is used for withdrawing assets from the layer-2 to the layer-1.

pubdata:

| Name               | Size(byte) | Comment                  |
| ------------------ | ---------- | ------------------------ |
| TxType             | 1          | transaction type         |
| AccountIndex       | 4          | from account index       |
| ToAddress          | 20         | layer-1 receiver address |
| AssetId            | 2          | asset index              |
| AssetAmount        | 16         | state amount             |
| GasFeeAccountIndex | 4          | gas fee account index    |
| GasFeeAssetId      | 2          | gas fee asset id         |
| GasFeeAssetAmount  | 2          | packed fee amount        |

### MintNft

This is a layer-2 transaction and is used for minting nfts in the layer-2 network.

pubdata:

| Name               | Size(byte) | Comment                |
| ------------------ | ---------- | ---------------------- |
| TxType             | 1          | transaction type       |
| FromAccountIndex   | 4          | from account index     |
| ToAccountIndex     | 4          | receiver account index |
| NftIndex           | 5          | unique nft index       |
| NftContentHash     | 32         | nft content hash       |
| GasFeeAccountIndex | 4          | gas fee account index  |
| GasFeeAssetId      | 2          | gas fee asset id       |
| GasFeeAssetAmount  | 2          | packed fee amount      |

### TransferNft

This is a layer-2 transaction and is used for transfering nfts to others in the layer-2 network.

pubdata:

| Name               | Size(byte) | Comment                |
| ------------------ | ---------- | ---------------------- |
| TxType             | 1          | transaction type       |
| FromAccountIndex   | 4          | from account index     |
| ToAccountIndex     | 4          | receiver account index |
| NftIndex           | 5          | unique nft index       |
| GasFeeAccountIndex | 4          | gas fee account index  |
| GasFeeAssetId      | 2          | gas fee asset id       |
| GasFeeAssetAmount  | 2          | packed fee amount      |
| CallDataHash       | 32         | call data hash         |

### SetNftPrice

This is a layer-2 transaction and is used for setting nft price in the layer-2 network.

pubdata:

| Name               | Size(byte) | Comment                |
| ------------------ | ---------- | ---------------------- |
| TxType             | 1          | transaction type       |
| FromAccountIndex   | 4          | from account index     |
| ToAccountIndex     | 4          | receiver account index |
| NftIndex           | 5          | unique nft index       |
| AssetId            | 2          | asset index            |
| AssetAmount        | 16         | state amount           |
| GasFeeAccountIndex | 4          | gas fee account index  |
| GasFeeAssetId      | 2          | gas fee asset id       |
| GasFeeAssetAmount  | 2          | packed fee amount      |

### BuyNft

This is a layer-2 transaction and is used for buying nfts in the layer-2 network.

pubdata:

| Name                    | Size(byte) | Comment                |
| ----------------------- | ---------- | ---------------------- |
| TxType                  | 1          | transaction type       |
| BuyerAccountIndex       | 4          | buyer account index    |
| OwnerAccountIndex       | 4          | owner account index    |
| NftIndex                | 5          | unique nft index       |
| AssetId                 | 2          | asset index            |
| AssetAmount             | 16         | state amount           |
| GasFeeAccountIndex      | 4          | gas fee account index  |
| GasFeeAssetId           | 2          | gas fee asset id       |
| GasFeeAssetAmount       | 2          | packed fee amount      |
| TreasuryFeeAccountIndex | 4          | treasury account index |
| TreasuryFeeAmount       | 2          | packed fee             |
| CreatorFeeAmount        | 2          | packed fee             |

### WithdrawNft

This is a layer-2 transaction and is used for withdrawing nft from the layer-2 to the layer-1.

pubdata:

| Name               | Size(byte) | Comment               |
| ------------------ | ---------- | --------------------- |
| TxType             | 1          | transaction type      |
| BuyerAccountIndex  | 4          | buyer account index   |
| OwnerAccountIndex  | 4          | owner account index   |
| NftIndex           | 5          | unique nft index      |
| NftContentHash     | 32         | nft content hash      |
| NftL1Address       | 20         | nft layer-1 address   |
| NftL1TokenId       | 32         | nft layer-1 token id  |
| AssetId            | 2          | asset index           |
| AssetAmount        | 16         | state amount          |
| GasFeeAccountIndex | 4          | gas fee account index |
| GasFeeAssetId      | 2          | gas fee asset id      |
| GasFeeAssetAmount  | 2          | packed fee amount     |

### FullExit

This is a layer-1 transaction and is used for full exit assets from the layer-2 to the layer-1.

pubdata:

| Name            | Size(byte) | Comment            |
| --------------- | ---------- | ------------------ |
| TxType          | 1          | transaction type   |
| AccountIndex    | 4          | from account index |
| AccountNameHash | 32         | account name hash  |
| AssetId         | 2          | asset index        |
| AssetAmount     | 16         | state amount       |

### FullExitNft

This is a layer-1 transaction and is used for full exit nfts from the layer-2 to the layer-1.

pubdata:

| Name           | Size(byte) | Comment              |
| -------------- | ---------- | -------------------- |
| TxType         | 1          | transaction type     |
| AccountIndex   | 4          | from account index   |
| NftIndex       | 5          | unique nft index     |
| NftContentHash | 32         | nft content hash     |
| NftL1Address   | 20         | nft layer-1 address  |
| NftL1TokenId   | 32         | nft layer-1 token id |