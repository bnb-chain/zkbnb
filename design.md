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

txType 1byte

### RegisterZNS

This is a layer-1 transaction and a user needs to call this method first to register a layer-2 account.

pubdata:

txType 1byte
accountName 32byte
pubKey 32byte

### CreatePair

This is a layer-1 transaction and is used for creating pair index for the layer-2 liquidity tree.

pubdata:

txType 1byte

pairIndex 2byte

assetAId 2byte

assetBId 2byte

### Deposit

This is a layer-1 transaction and is used for depositing assets into the layer-2 account.

pubdata:

txType 1byte
accountIndex 4byte
accountNameHash 32byte
assetId 2byte
assetAmount 16byte

### DepositNft

This is a layer-1 transaction and is used for depositing nfts into the layer-2 account.

pubdata:

txType 1byte
accountIndex 4byte
accountNameHash 32byte
nftIndex 5byte
nftContentHash 32byte
nftL1Address 20byte
nftL1TokenId 32byte

### Transfer

This is a layer-2 transaction and is used for transfering assets in the layer-2 network.

pubdata:

txType 1byte
fromAccountIndex 4byte
toAccountIndex 4byte
assetId 2byte
assetAmount 5byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte
callDataHash 32byte

### Swap

This is a layer-2 transaction and is used for making a swap for assets in the layer-2 network.

pubdata:

txType 1byte
fromAccountIndex 4byte
toAccountIndex 4byte
pairIndex 2byte
assetAAmount 5byte
assetBAmount 5byte
treasuryAccountIndex 4byte
treasuryFeeAmount 2byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte

### AddLiquidity

This is a layer-2 transaction and is used for adding liquidity for a trading pair in the layer-2 network.

pubdata:

txType 1byte
fromAccountIndex 4byte
toAccountIndex 4byte
pairIndex 2byte
assetAAmount 5byte
assetBAmount 5byte
lpAmount 5byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte

### RemoveLiquidity

This is a layer-2 transaction and is used for removing liquidity for a trading pair in the layer-2 network.

pubdata:

txType 1byte
fromAccountIndex 4byte
toAccountIndex 4byte
pairIndex 2byte
assetAAmount 5byte
assetBAmount 5byte
lpAmount 5byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte

### Withdraw

This is a layer-2 transaction and is used for withdrawing assets from the layer-2 to the layer-1.

pubdata:

txType 1byte
accountIndex 4byte
toAddress 20byte
assetId 2byte
assetAmount 16byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte

### MintNft

This is a layer-2 transaction and is used for minting nfts in the layer-2 network.

pubdata:

txType 1byte
fromAccountIndex 4byte
toAccountIndex 4byte
nftIndex 5byte
nftContentHash 32byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte

### TransferNft

This is a layer-2 transaction and is used for transfering nfts to others in the layer-2 network.

pubdata:

txType 1byte
fromAccountIndex 4byte
toAccountIndex 4byte
nftIndex 5byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte
callDataHash 32byte

### SetNftPrice

This is a layer-2 transaction and is used for setting nft price in the layer-2 network.

pubdata:

txType 1byte
accountIndex 4byte
nftIndex 5byte
assetId 2byte
assetAmount 5byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte

### BuyNft

This is a layer-2 transaction and is used for buying nfts in the layer-2 network.

pubdata:

txType 1byte
buyerAccountIndex 4byte
ownerAccountIndex 4byte
nftIndex 5byte
assetId 2byte
assetAmount 5byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte
treasuryFeeAccountIndex 4byte
treasuryFeeAmount 2byte
creatorFeeAmount 2byte

### WithdrawNft

This is a layer-2 transaction and is used for withdrawing nft from the layer-2 to the layer-1.

pubdata:

txType 1byte
accountIndex 4byte
toAddress 20byte
proxyAddress 20byte
nftIndex 5byte
nftContentHash 32byte
nftL1Address 20byte
nftL1TokenId 32byte
gasFeeAccountIndex 4byte
gasFeeAssetId 2byte
gasFeeAssetAmount 2byte

### FullExit

This is a layer-1 transaction and is used for full exit assets from the layer-2 to the layer-1.

pubdata:

txType 1byte
accountIndex 4byte
accountNameHash 32byte
assetId 2byte
fullAmount 16byte

### FullExitNft

This is a layer-1 transaction and is used for full exit nfts from the layer-2 to the layer-1.

pubdata:

txType 1byte
accountIndex 4byte
accountNameHash 32byte
nftIndex 5byte
nftContentHash 32byte
nftL1Address 20byte
nftL1TokenId 32byte