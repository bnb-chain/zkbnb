# ZkBNB Protocol Design

## Glossary

- **L1**: layer 1 blockchain, it is BNB Smart Chain.
- **Rollup**: Zk Rollup based layer-2 network, it is ZkBNB.
- **Owner**: A user get a L2 account.
- **Committer**: Entity executing transactions and producing consecutive blocks on L2.
- **Eventually**: happening within finite time.
- **Assets in L2**: Assets in L2 smart contract controlled by owners.
- **L2 Key**: Owner's private key used to send transaction on L2.
- **MiMC Signature**: The result of signing the owner's message, 
using his private key, used in L2 internal transactions.

The current implementation we still use EDDSA as the signature scheme, we will soon support
switch to EDCSA.

## Design

### Overview

ZkBNB implements a ZK rollup protocol (in short "rollup" below) for:

- BNB and BEP20 fungible token deposit and transfer
- AMM-based fungible token swap on L2
- BEP721 non-fungible token deposit and transfer
- mint BEP721 non-fungible tokens on L2
- NFT-marketplace on L2

General rollup workflow is as follows:

- Users can become owners in rollup by calling registerZNS in L1 to register a short name for L2;
- Owners can transfer assets to each other, mint NFT on L2 or make a swap on L2;
- Owners can withdraw assets under their control to any L1 address.

Rollup operation requires the assistance of a committer, who rolls transactions together, also a prover who computes
a zero-knowledge proof of the correct state transition, and affects the state transition by interacting with the 
rollup contract.

## Data format

### Data types

We assume that 1 `Chunk` = 32 bytes.

| Type           | Size(Byte) | Type     | Comment                                                                                                                                                                                                                                                                                                              |
|----------------|------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AccountIndex   | 4          | uint32   | Incremented number of accounts in Rollup. New account will have the next free id. Max value is 2^32 - 1 = 4.294967295 × 10^9                                                                                                                                                                                         |
| AssetId        | 2          | uint16   | Incremented number of tokens in Rollup, max value is 65535                                                                                                                                                                                                                                                           |
| PackedTxAmount | 5          | int64    | Packed transactions amounts are represented with 40 bit (5 byte) values, encoded as mantissa × 10^exponent where mantissa is represented with 35 bits, exponent is represented with 5 bits. This gives a range from 0 to 34359738368 × 10^31, providing 10 full decimal digit precision.                             |
| PackedFee      | 2          | uint16   | Packed fees must be represented with 2 bytes: 5 bit for exponent, 11 bit for mantissa.                                                                                                                                                                                                                               |
| StateAmount    | 16         | *big.Int | State amount is represented as uint128 with a range from 0 to ~3.4 × 10^38. It allows to represent up to 3.4 × 10^20 "units" if standard Ethereum's 18 decimal symbols are used. This should be a sufficient range.                                                                                                  |
| Nonce          | 4          | uint32   | Nonce is the total number of executed transactions of the account. In order to apply the update of this state, it is necessary to indicate the current account nonce in the corresponding transaction, after which it will be automatically incremented. If you specify the wrong nonce, the changes will not occur. |
| EthAddress     | 20         | string   | To make an BNB Smart Chain address from the BNB Smart Chain's public key, all we need to do is to apply Keccak-256 hash function to the key and then take the last 20 bytes of the result.                                                                                                                           |
| Signature      | 64         | []byte   | Based on EDDSA.                                                                                                                                                                                                                                                                                                      |
| HashValue      | 32         | string   | hash value based on MiMC                                                                                                                                                                                                                                                                                             |

### Amount packing

Mantissa and exponent parameters used in ZkBNB:

`amount = mantissa * radix^{exponent}`

| Type           | Exponent bit width | Mantissa bit width | Radix |
|----------------|--------------------|--------------------|-------|
| PackedTxAmount | 5                  | 35                 | 10    |
| PackedFee      | 5                  | 11                 | 10    |

### State Merkle Tree(height)

We have 3 unique trees: `AccountTree(32)`, `LiquidityTree(16)`, `NftTree(40)` and one sub-tree `AssetTree(16)` which 
belongs to `AccountTree(32)`. The empty leaf for all the trees is just set every attribute as `0` for every node.

#### AccountTree

`AccountTree` is used for storing all accounts info and the node of the account tree is:

```go
type AccountNode struct{
    AccountNameHash string // bytes32
    PubKey string // bytes32
    Nonce int64
    CollectionNonce int64
    AssetRoot string // bytes32
}
```

Leaf hash computation:

```go
func ComputeAccountLeafHash(
	accountNameHash string,
	pk string,
	nonce int64,
	collectionNonce int64,
	assetRoot []byte,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	buf.Write(common.FromHex(accountNameHash))
	err = util.PaddingPkIntoBuf(&buf, pk)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] unable to write pk into buf: %s", err.Error())
		return nil, err
	}
	util.PaddingInt64IntoBuf(&buf, nonce)
	util.PaddingInt64IntoBuf(&buf, collectionNonce)
	buf.Write(assetRoot)
	hFunc.Reset()
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}
```

##### AssetTree

`AssetTree` is sub-tree of `AccountTree` and it stores all the assets `balance`, `lpAmount` and `offerCanceledOrFinalized`. The node of asset tree is:

```go
type AssetNode struct {
	Balance  string
	LpAmount string
    OfferCanceledOrFinalized string // uint128
}
```

Leaf hash computation:

```go
func ComputeAccountAssetLeafHash(
	balance string,
	lpAmount string,
	offerCanceledOrFinalized string,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	err = util.PaddingStringBigIntIntoBuf(&buf, balance)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	err = util.PaddingStringBigIntIntoBuf(&buf, lpAmount)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	err = util.PaddingStringBigIntIntoBuf(&buf, offerCanceledOrFinalized)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	hFunc.Write(buf.Bytes())
	return hFunc.Sum(nil), nil
}
```

#### LiquidityTree

`LiquidityTree` is used for storing all the liquidity info and the node of the liquidity tree is:

```go
type LiquidityNode struct {
    AssetAId             int64
	AssetA               string
	AssetBId             int64
	AssetB               string
	LpAmount             string
	KLast                string
	FeeRate              int64
	TreasuryAccountIndex int64
	TreasuryRate         int64
}
```

The liquidity pair is first initialized by `CreatePair` tx and will be changed by `UpdatePairRate`, `Swap`, `AddLiquidity` and `RemoveLiquidity` txs.

Leaf hash computation:

```go
func ComputeLiquidityAssetLeafHash(
	assetAId int64,
	assetA string,
	assetBId int64,
	assetB string,
	lpAmount string,
	kLast string,
	feeRate int64,
	treasuryAccountIndex int64,
	treasuryRate int64,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	util.PaddingInt64IntoBuf(&buf, assetAId)
	err = util.PaddingStringBigIntIntoBuf(&buf, assetA)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	util.PaddingInt64IntoBuf(&buf, assetBId)
	err = util.PaddingStringBigIntIntoBuf(&buf, assetB)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	err = util.PaddingStringBigIntIntoBuf(&buf, lpAmount)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	err = util.PaddingStringBigIntIntoBuf(&buf, kLast)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	util.PaddingInt64IntoBuf(&buf, feeRate)
	util.PaddingInt64IntoBuf(&buf, treasuryAccountIndex)
	util.PaddingInt64IntoBuf(&buf, treasuryRate)
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}
```

#### NftTree

`NftTree` is used for storing all the NFTs and the node info is:

```go
type NftNode struct {
    CreatorAccountIndex int64
    OwnerAccountIndex   int64
    NftContentHash      string
    NftL1Address        string
    NftL1TokenId        string
    CreatorTreasuryRate int64
    CollectionId        int64 // 32 bit
}
```

Leaf hash computation:

```go
func ComputeNftAssetLeafHash(
	creatorAccountIndex int64,
	ownerAccountIndex int64,
	nftContentHash string,
	nftL1Address string,
	nftL1TokenId string,
	creatorTreasuryRate int64,
	collectionId int64,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	util.PaddingInt64IntoBuf(&buf, creatorAccountIndex)
	util.PaddingInt64IntoBuf(&buf, ownerAccountIndex)
	buf.Write(ffmath.Mod(new(big.Int).SetBytes(common.FromHex(nftContentHash)), curve.Modulus).FillBytes(make([]byte, 32)))
	err = util.PaddingAddressIntoBuf(&buf, nftL1Address)
	if err != nil {
		logx.Errorf("[ComputeNftAssetLeafHash] unable to write address to buf: %s", err.Error())
		return nil, err
	}
	err = util.PaddingStringBigIntIntoBuf(&buf, nftL1TokenId)
	if err != nil {
		logx.Errorf("[ComputeNftAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	util.PaddingInt64IntoBuf(&buf, creatorTreasuryRate)
	util.PaddingInt64IntoBuf(&buf, collectionId)
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}
```

#### StateRoot

`StateRoot` is the final root that shows the final layer-2 state and will be stored on L1. It is computed by the root of `AccountTree`, `LiquidityTree` and `NftTree`. The computation of `StateRoot` works as follows:

`StateRoot = MiMC(AccountRoot || LiquidityRoot || NftRoot)`

```go
func ComputeStateRootHash(
	accountRoot []byte,
	liquidityRoot []byte,
	nftRoot []byte,
) []byte {
	hFunc := mimc.NewMiMC()
	hFunc.Write(accountRoot)
	hFunc.Write(liquidityRoot)
	hFunc.Write(nftRoot)
	return hFunc.Sum(nil)
}
```

## ZkBNB Transactions

ZkBNB transactions are divided into Rollup transactions (initiated inside Rollup by a Rollup account) and Priority operations (initiated on the BSC by an BNB Smart Chain account).

Rollup transactions:

- EmptyTx
- Transfer
- Swap
- AddLiquidity
- RemoveLiquidity
- Withdraw
- CreateCollection
- MintNft
- TransferNft
- AtomicMatch
- CancelOffer
- WithdrawNft

Priority operations:

- RegisterZNS
- CreatePair
- UpdatePairRate
- Deposit
- DepositNft
- FullExit
- FullExitNft

### Rollup transaction lifecycle

1. User creates a `Transaction` or a `Priority operation`.
2. After processing this request, committer creates a `Rollup operation` and adds it to the block.
3. Once the block is complete, sender submits it to the ZkBNB smart contract as a block commitment.
   Part of the logic of some `Rollup transaction` is checked by the smart contract.
4. The proof for the block is submitted to the ZkBNB smart contract as the block verification.
   If the verification succeeds, the new state is considered finalized.

### EmptyTx

#### Description

No effects.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
| 1      | 1                 |

##### Structure

| Field  | Size(byte) | Value/type | Description      |
|--------|------------|------------|------------------|
| TxType | 1          | `0x00`     | Transaction type |

#### User transaction

No user transaction

### RegisterZNS

#### Description

This is a layer-1 transaction and a user needs to call this method first to register a layer-2 account.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 101               |

##### Structure

| Name            | Size(byte) | Comment                        |
|-----------------|------------|--------------------------------|
| TxType          | 1          | transaction type               |
| AccountIndex    | 4          | unique account index           |
| AccountName     | 32         | account name                   |
| AccountNameHash | 32         | hash value of the account name |
| PubKeyX         | 32         | layer-2 account's public key X |
| PubKeyY         | 32         | layer-2 account's public key Y |

```go
func ConvertTxToRegisterZNSPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeRegisterZns {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToRegisterZNSPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseRegisterZnsTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(AccountNameToBytes32(txInfo.AccountName)))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	pk, err := ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] unable to parse pub key: %s", err.Error())
		return nil, err
	}
	// because we can get Y from X, so we only need to store X is enough
	buf.Write(PrefixPaddingBufToChunkSize(pk.A.X.Marshal()))
	buf.Write(PrefixPaddingBufToChunkSize(pk.A.Y.Marshal()))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

| Name        | Size(byte) | Comment                      |
| ----------- | ---------- | ---------------------------- |
| AccountName | 32         | account name                 |
| Owner       | 20         | account layer-1 address      |
| PubKey      | 32         | layer-2 account's public key |

#### Circuit

```go
func VerifyRegisterZNSTx(
	api API, flag Variable,
	tx RegisterZnsTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromRegisterZNS(api, tx)
	CheckEmptyAccountNode(api, flag, accountsBefore[0])
	return pubData
}
```

### CreatePair

#### Description

This is a layer-1 transaction and is used for creating a trading pair for L2.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 15                |

##### Structure

| Name                 | Size(byte) | Comment                       |
| -------------------- | ---------- | ----------------------------- |
| TxType               | 1          | transaction type              |
| PairIndex            | 2          | unique pair index             |
| AssetAId             | 2          | unique asset index            |
| AssetBId             | 2          | unique asset index            |
| FeeRate              | 2          | fee rate                      |
| TreasuryAccountIndex | 4          | unique treasury account index |
| TreasuryRate         | 2          | treasury rate                 |

```go
func ConvertTxToCreatePairPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeCreatePair {
		logx.Errorf("[ConvertTxToCreatePairPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToCreatePairPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseCreatePairTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToCreatePairPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetAId)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetBId)))
	buf.Write(Uint16ToBytes(uint16(txInfo.FeeRate)))
	buf.Write(Uint32ToBytes(uint32(txInfo.TreasuryAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.TreasuryRate)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

| Name          | Size(byte) | Comment                 |
| ------------- | ---------- | ----------------------- |
| AssetAAddress | 20         | asset a layer-1 address |
| AssetBAddress | 20         | asset b layer-1 address |

#### Circuit

```go
func VerifyCreatePairTx(
	api API, flag Variable,
	tx CreatePairTxConstraints,
	liquidityBefore LiquidityConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromCreatePair(api, tx)
	// verify params
	IsVariableEqual(api, flag, tx.PairIndex, liquidityBefore.PairIndex)
	CheckEmptyLiquidityNode(api, flag, liquidityBefore)
	return pubData
}
```

### UpdatePairRate

#### Description

This is a layer-1 transaction and is used for updating a trading pair for L2.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 11                |

##### Structure

| Name                 | Size(byte) | Comment                       |
| -------------------- | ---------- | ----------------------------- |
| TxType               | 1          | transaction type              |
| PairIndex            | 2          | unique pair index             |
| FeeRate              | 2          | fee rate                      |
| TreasuryAccountIndex | 4          | unique treasury account index |
| TreasuryRate         | 2          | treasury rate                 |

```go
func ConvertTxToUpdatePairRatePubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeUpdatePairRate {
		logx.Errorf("[ConvertTxToUpdatePairRatePubData] invalid tx type")
		return nil, errors.New("[ConvertTxToUpdatePairRatePubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseUpdatePairRateTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToUpdatePairRatePubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.FeeRate)))
	buf.Write(Uint32ToBytes(uint32(txInfo.TreasuryAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.TreasuryRate)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

| Name                 | Size(byte) | Comment                 |
| -------------------- | ---------- | ----------------------- |
| AssetAAddress        | 20         | asset a layer-1 address |
| AssetBAddress        | 20         | asset b layer-1 address |
| FeeRate              | 2          | fee rate                |
| TreasuryAccountIndex | 4          | treasury account index  |
| TreasuryRate         | 2          | treasury rate           |

#### Circuit

```go
func VerifyUpdatePairRateTx(
	api API, flag Variable,
	tx UpdatePairRateTxConstraints,
	liquidityBefore LiquidityConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromUpdatePairRate(api, tx)
	// verify params
	IsVariableEqual(api, flag, tx.PairIndex, liquidityBefore.PairIndex)
	IsVariableLessOrEqual(api, flag, tx.TreasuryRate, tx.FeeRate)
	return pubData
}
```

### Deposit

#### Description

This is a layer-1 transaction and is used for depositing assets into the layer-2 account.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 55                |

##### Structure

| Name            | Size(byte) | Comment           |
| --------------- | ---------- | ----------------- |
| TxType          | 1          | transaction type  |
| AccountIndex    | 4          | account index     |
| AssetId         | 2          | asset index       |
| AssetAmount     | 16         | state amount      |
| AccountNameHash | 32         | account name hash |

```go
func ConvertTxToDepositPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeDeposit {
		logx.Errorf("[ConvertTxToDepositPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToDepositPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseDepositTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

##### DepositBNB

| Name            | Size(byte) | Comment           |
| --------------- | ---------- | ----------------- |
| AccountNameHash | 32         | account name hash |

##### DepositBEP20

| Name            | Size(byte) | Comment               |
| --------------- | ---------- | --------------------- |
| AssetAddress    | 20         | asset layer-1 address |
| Amount          | 13         | asset layer-1 amount  |
| AccountNameHash | 32         | account name hash     |

#### Circuit

```go
func VerifyDepositTx(
	api API, flag Variable,
	tx DepositTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromDeposit(api, tx)
	// verify params
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	return pubData
}
```

### DepositNft

#### Description

This is a layer-1 transaction and is used for depositing NFTs into the layer-2 account.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 134               |

##### Structure

| Name                | Size(byte) | Comment               |
| ------------------- | ---------- | --------------------- |
| TxType              | 1          | transaction type      |
| AccountIndex        | 4          | account index         |
| NftIndex            | 5          | unique index of a nft |
| NftL1Address        | 20         | nft layer-1 address   |
| CreatorAccountIndex | 4          | creator account index |
| CreatorTreasuryRate | 2          | creator treasury rate |
| CollectionId        | 2          | collection id         |
| NftContentHash      | 32         | nft content hash      |
| NftL1TokenId        | 32         | nft layer-1 token id  |
| AccountNameHash     | 32         | account name hash     |

```go
func ConvertTxToDepositNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeDepositNft {
		logx.Errorf("[ConvertTxToDepositNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToDepositNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseDepositNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

| Name            | Size(byte) | Comment                      |
| --------------- | ---------- | ---------------------------- |
| AccountNameHash | 32         | account name hash            |
| AssetAddress    | 20         | nft contract layer-1 address |
| NftTokenId      | 32         | nft layer-1 token id         |

#### Circuit

```go
func VerifyDepositNftTx(
	api API,
	flag Variable,
	tx DepositNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromDepositNft(api, tx)
	// verify params
	// check empty nft
	CheckEmptyNftNode(api, flag, nftBefore)
	// account index
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	// account name hash
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
	return pubData
}
```

### Transfer

#### Description

This is a layer-2 transaction and is used for transferring assets in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 56                |

##### Structure

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

```go
func ConvertTxToTransferPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeTransfer {
		logx.Errorf("[ConvertTxToTransferPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToTransferPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseTransferTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	pubData = buf.Bytes()
	return pubData, nil
}
```

#### User transaction

```go
type TransferTxInfo struct {
	FromAccountIndex  int64
	ToAccountIndex    int64
	ToAccountNameHash string
	AssetId           int64
	AssetAmount       *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	Memo              string
	CallData          string
	CallDataHash      []byte
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyTransferTx(
	api API, flag Variable,
	tx *TransferTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	// collect pub-data
	pubData = CollectPubDataFromTransfer(api, *tx)
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.ToAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// account name hash
	IsVariableEqual(api, flag, tx.ToAccountNameHash, accountsBefore[1].AccountNameHash)
	// asset id
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// gas asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	// should have enough balance
	tx.AssetAmount = UnpackAmount(api, tx.AssetAmount)
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	//tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[1].Balance)
	return pubData
}
```

### Swap

#### Description

This is a layer-2 transaction and is used for making a swap for assets in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 25                |

##### Structure

| Name               | Size(byte) | Comment               |
| ------------------ | ---------- | --------------------- |
| TxType             | 1          | transaction type      |
| FromAccountIndex   | 4          | from account index    |
| PairIndex          | 2          | unique pair index     |
| AssetAAmount       | 5          | packed asset amount   |
| AssetBAmount       | 5          | packed asset amount   |
| GasFeeAccountIndex | 4          | gas fee account index |
| GasFeeAssetId      | 2          | gas fee asset id      |
| GasFeeAssetAmount  | 2          | packed fee amount     |

```go
func ConvertTxToSwapPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeSwap {
		logx.Errorf("[ConvertTxToSwapPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToSwapPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseSwapTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToSwapPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountDeltaBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountDeltaBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type SwapTxInfo struct {
	FromAccountIndex  int64
	PairIndex         int64
	AssetAId          int64
	AssetAAmount      *big.Int
	AssetBId          int64
	AssetBMinAmount   *big.Int
	AssetBAmountDelta *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifySwapTx(
	api API, flag Variable,
	tx *SwapTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints, liquidityBefore LiquidityConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromSwap(api, *tx)
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// pair index
	IsVariableEqual(api, flag, tx.PairIndex, liquidityBefore.PairIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetAId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetBId, accountsBefore[0].AssetsInfo[1].AssetId)
	isSameAsset := api.IsZero(
		api.And(
			api.IsZero(api.Sub(tx.AssetAId, liquidityBefore.AssetAId)),
			api.IsZero(api.Sub(tx.AssetBId, liquidityBefore.AssetBId)),
		),
	)
	isDifferentAsset := api.IsZero(
		api.And(
			api.IsZero(api.Sub(tx.AssetAId, liquidityBefore.AssetBId)),
			api.IsZero(api.Sub(tx.AssetBId, liquidityBefore.AssetAId)),
		),
	)
	IsVariableEqual(
		api, flag,
		api.Or(
			isSameAsset,
			isDifferentAsset,
		),
		1,
	)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[2].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// should have enough assets
	tx.AssetAAmount = UnpackAmount(api, tx.AssetAAmount)
	tx.AssetBMinAmount = UnpackAmount(api, tx.AssetBMinAmount)
	tx.AssetBAmountDelta = UnpackAmount(api, tx.AssetBAmountDelta)
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.AssetBMinAmount, tx.AssetBAmountDelta)
	IsVariableLessOrEqual(api, flag, tx.AssetAAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[2].Balance)
	// pool info
	isSameAsset = api.And(flag, isSameAsset)
	isDifferentAsset = api.And(flag, isSameAsset)
	IsVariableEqual(api, flag, liquidityBefore.FeeRate, liquidityBefore.FeeRate)
	IsVariableLessOrEqual(api, flag, liquidityBefore.FeeRate, RateBase)
	assetAAmount := api.Select(isSameAsset, tx.AssetAAmount, tx.AssetBAmountDelta)
	assetBAmount := api.Select(isSameAsset, tx.AssetBAmountDelta, tx.AssetAAmount)
	// verify AMM
	r := api.Mul(api.Mul(liquidityBefore.AssetA, liquidityBefore.AssetB), RateBase)
	l := api.Mul(
		api.Sub(
			api.Mul(RateBase, api.Add(assetAAmount, liquidityBefore.AssetA)),
			api.Mul(liquidityBefore.FeeRate, assetAAmount),
		),
		api.Add(assetBAmount, liquidityBefore.AssetB),
	)
	IsVariableLessOrEqual(api, flag, r, l)
	return pubData
}
```

### AddLiquidity

#### Description

This is a layer-2 transaction and is used for adding liquidity for a trading pair in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 40                |

##### Structure

| Name               | Size(byte) | Comment                |
| ------------------ | ---------- | ---------------------- |
| TxType             | 1          | transaction type       |
| FromAccountIndex   | 4          | from account index     |
| PairIndex          | 2          | unique pair index      |
| AssetAAmount       | 5          | packed asset amount    |
| AssetBAmount       | 5          | packed asset amount    |
| LpAmount           | 5          | packed asset amount    |
| KLast              | 5          | packed k last amount   |
| TreasuryAmount     | 5          | packed treasury amount |
| GasFeeAccountIndex | 4          | gas fee account index  |
| GasFeeAssetId      | 2          | gas fee asset id       |
| GasFeeAssetAmount  | 2          | packed fee amount      |

```go
func ConvertTxToAddLiquidityPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeAddLiquidity {
		logx.Errorf("[ConvertTxToAddLiquidityPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToAddLiquidityPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToAddLiquidityPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountBytes)
	LpAmountBytes, err := AmountToPackedAmountBytes(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(LpAmountBytes)
	KLastBytes, err := AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(KLastBytes)
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	treasuryAmountBytes, err := AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type AddLiquidityTxInfo struct {
	FromAccountIndex  int64
	PairIndex         int64
	AssetAId          int64
	AssetAAmount      *big.Int
	AssetBId          int64
	AssetBAmount      *big.Int
	LpAmount          *big.Int
	KLast             *big.Int
	TreasuryAmount    *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyAddLiquidityTx(
	api API, flag Variable,
	tx *AddLiquidityTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints, liquidityBefore LiquidityConstraints,
	hFunc *MiMC,
) {
	CollectPubDataFromAddLiquidity(api, flag, *tx, hFunc)
	// check params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, liquidityBefore.TreasuryAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetAId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetBId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.AssetAId, liquidityBefore.AssetAId)
	IsVariableEqual(api, flag, tx.AssetBId, liquidityBefore.AssetBId)
	IsVariableEqual(api, flag, tx.PairIndex, accountsBefore[1].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[2].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	IsVariableLessOrEqual(api, flag, 0, tx.AssetAAmount)
	IsVariableLessOrEqual(api, flag, 0, tx.AssetBAmount)
	// check if the user has enough balance
	tx.AssetAAmount = UnpackAmount(api, tx.AssetAAmount)
	tx.AssetBAmount = UnpackAmount(api, tx.AssetBAmount)
	tx.LpAmount = UnpackAmount(api, tx.LpAmount)
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.AssetAAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.AssetBAmount, accountsBefore[0].AssetsInfo[1].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[2].Balance)
	IsVariableEqual(api, flag, tx.PoolAAmount, liquidityBefore.AssetA)
	IsVariableEqual(api, flag, tx.PoolBAmount, liquidityBefore.AssetB)
	// TODO verify ratio
	l := api.Mul(liquidityBefore.AssetA, tx.AssetAAmount)
	r := api.Mul(liquidityBefore.AssetB, tx.AssetBAmount)
	maxDelta := std.Max(api, liquidityBefore.AssetA, liquidityBefore.AssetB)
	l = std.Max(api, l, r)
	r = std.Min(api, l, r)
	lrDelta := api.Sub(l, r)
	IsVariableLessOrEqual(api, flag, lrDelta, maxDelta)
	// TODO verify lp amount
	zeroFlag := api.Compiler().IsBoolean(api.Add(liquidityBefore.AssetA, 1))
	if zeroFlag {
		// lpAmount = \sqrt{x * y}
		lpAmountSquare := api.Mul(tx.AssetAAmount, tx.AssetBAmount)
		IsVariableEqual(api, flag, api.Mul(tx.LpAmount, tx.LpAmount), lpAmountSquare)
	} else {
		// lpAmount = \Delta{x} / x * poolLp
		IsVariableEqual(api, flag, api.Mul(tx.LpAmount, liquidityBefore.AssetA), api.Mul(tx.AssetAAmount, liquidityBefore.LpAmount))
	}
}
```

### RemoveLiquidity

#### Description

This is a layer-2 transaction and is used for removing liquidity for a trading pair in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 40                |

##### Structure

| Name               | Size(byte) | Comment                |
| ------------------ | ---------- | ---------------------- |
| TxType             | 1          | transaction type       |
| FromAccountIndex   | 4          | from account index     |
| PairIndex          | 2          | unique pair index      |
| AssetAAmount       | 5          | packed asset amount    |
| AssetBAmount       | 5          | packed asset amount    |
| LpAmount           | 5          | packed asset amount    |
| KLast              | 5          | packed k last amount   |
| TreasuryAmount     | 5          | packed treasury amount |
| GasFeeAccountIndex | 4          | gas fee account index  |
| GasFeeAssetId      | 2          | gas fee asset id       |
| GasFeeAssetAmount  | 2          | packed fee amount      |

```go
func ConvertTxToRemoveLiquidityPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeRemoveLiquidity {
		logx.Errorf("[ConvertTxToRemoveLiquidityPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToRemoveLiquidityPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseRemoveLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToRemoveLiquidityPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountBytes)
	LpAmountBytes, err := AmountToPackedAmountBytes(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(LpAmountBytes)
	KLastBytes, err := AmountToPackedAmountBytes(txInfo.KLast)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(KLastBytes)
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	treasuryAmountBytes, err := AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type RemoveLiquidityTxInfo struct {
	FromAccountIndex  int64
	PairIndex         int64
	AssetAId          int64
	AssetAMinAmount   *big.Int
	AssetBId          int64
	AssetBMinAmount   *big.Int
	LpAmount          *big.Int
	AssetAAmountDelta *big.Int
	AssetBAmountDelta *big.Int
	KLast             *big.Int
	TreasuryAmount    *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyRemoveLiquidityTx(
	api API, flag Variable,
	tx *RemoveLiquidityTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints, liquidityBefore LiquidityConstraints,
) (pubData [PubDataSizePerTx]Variable, err error) {
	pubData = CollectPubDataFromRemoveLiquidity(api, *tx)
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, liquidityBefore.TreasuryAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetAId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetBId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[2].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetAId, liquidityBefore.AssetAId)
	IsVariableEqual(api, flag, tx.AssetBId, liquidityBefore.AssetBId)
	// should have enough lp
	IsVariableLessOrEqual(api, flag, tx.LpAmount, accountsBefore[0].AssetsInfo[3].LpAmount)
	// enough balance
	tx.AssetAMinAmount = UnpackAmount(api, tx.AssetAMinAmount)
	tx.AssetAAmountDelta = UnpackAmount(api, tx.AssetAAmountDelta)
	tx.AssetBMinAmount = UnpackAmount(api, tx.AssetBMinAmount)
	tx.AssetBAmountDelta = UnpackAmount(api, tx.AssetBAmountDelta)
	tx.LpAmount = UnpackAmount(api, tx.LpAmount)
	tx.TreasuryAmount = UnpackAmount(api, tx.TreasuryAmount)
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[3].Balance)
	// TODO verify LP
	kCurrent := api.Mul(liquidityBefore.AssetA, liquidityBefore.AssetB)
	IsVariableLessOrEqual(api, flag, liquidityBefore.KLast, kCurrent)
	IsVariableLessOrEqual(api, flag, liquidityBefore.TreasuryRate, liquidityBefore.FeeRate)
	sLps, err := api.Compiler().NewHint(ComputeSLp, 1, liquidityBefore.AssetA, liquidityBefore.AssetB, liquidityBefore.KLast, liquidityBefore.FeeRate, liquidityBefore.TreasuryRate)
	if err != nil {
		return pubData, err
	}
	sLp := sLps[0]
	IsVariableEqual(api, flag, tx.TreasuryAmount, sLp)
	poolLpVar := api.Sub(liquidityBefore.LpAmount, sLp)
	IsVariableLessOrEqual(api, flag, api.Mul(tx.AssetAAmountDelta, poolLpVar), api.Mul(tx.LpAmount, liquidityBefore.AssetA))
	IsVariableLessOrEqual(api, flag, api.Mul(tx.AssetBAmountDelta, poolLpVar), api.Mul(tx.LpAmount, liquidityBefore.AssetB))
	IsVariableLessOrEqual(api, flag, tx.AssetAMinAmount, tx.AssetAAmountDelta)
	IsVariableLessOrEqual(api, flag, tx.AssetBMinAmount, tx.AssetBAmountDelta)
	return pubData, nil
}
```

### Withdraw

#### Description

This is a layer-2 transaction and is used for withdrawing assets from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 51                |

##### Structure

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

```go
func ConvertTxToWithdrawPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeWithdraw {
		logx.Errorf("[ConvertTxToWithdrawPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToWithdrawPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseWithdrawTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToWithdrawPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(AddressStrToBytes(txInfo.ToAddress))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type WithdrawTxInfo struct {
	FromAccountIndex  int64
	AssetId           int64
	AssetAmount       *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ToAddress         string
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyWithdrawTx(
	api API, flag Variable,
	tx *WithdrawTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromWithdraw(api, *tx)
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// should have enough assets
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[1].Balance)
	return pubData
}
```

### CreateCollection

#### Description

This is a layer-2 transaction and is used for creating a new collection

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 15                |

##### Structure

| Name              | Size(byte) | Comment           |
| ----------------- | ---------- | ----------------- |
| TxType            | 1          | transaction type  |
| AccountIndex      | 4          | account index     |
| CollectionId      | 2          | collection index  |
| GasAccountIndex   | 4          | gas account index |
| GasFeeAssetId     | 2          | asset id          |
| GasFeeAssetAmount | 2          | packed fee amount |

```go
func ConvertTxToCreateCollectionPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeCreateCollection {
		logx.Errorf("[ConvertTxToCreateCollectionPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToCreateCollectionPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseCreateCollectionTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToCreateCollectionPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type CreateCollectionTxInfo struct {
	AccountIndex      int64
	CollectionId      int64
	Name              string
	Introduction      string
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyCreateCollectionTx(
	api API, flag Variable,
	tx *CreateCollectionTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromCreateCollection(api, *tx)
	// verify params
	IsVariableLessOrEqual(api, flag, tx.CollectionId, 65535)
	// account index
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// collection id
	IsVariableEqual(api, flag, tx.CollectionId, api.Add(accountsBefore[0].CollectionNonce, 1))
	// should have enough assets
	tx.GasFeeAssetAmount = UnpackAmount(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	return pubData
}
```

### MintNft

#### Description

This is a layer-2 transaction and is used for minting NFTs in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 58                |

##### Structure

| Name                | Size(byte) | Comment                |
|---------------------| ---------- | ---------------------- |
| TxType              | 1          | transaction type       |
| CreatorAccountIndex | 4          | creator account index  |
| ToAccountIndex      | 4          | receiver account index |
| NftIndex            | 5          | unique nft index       |
| GasFeeAccountIndex  | 4          | gas fee account index  |
| GasFeeAssetId       | 2          | gas fee asset id       |
| GasFeeAssetAmount   | 2          | packed fee amount      |
| CreatorTreasuryRate | 2          | creator treasury rate  |
| CollectionId        | 2          | collection index       |
| NftContentHash      | 32         | nft content hash       |

```go
func ConvertTxToMintNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeMintNft {
		logx.Errorf("[ConvertTxToMintNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToMintNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseMintNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToMintNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint16ToBytes(uint16(txInfo.NftCollectionId)))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(common.FromHex(txInfo.NftContentHash)))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type MintNftTxInfo struct {
	CreatorAccountIndex int64
	ToAccountIndex      int64
	ToAccountNameHash   string
	NftIndex            int64
	NftContentHash      string
	NftCollectionId     int64
	CreatorTreasuryRate int64
	GasAccountIndex     int64
	GasFeeAssetId       int64
	GasFeeAssetAmount   *big.Int
	ExpiredAt           int64
	Nonce               int64
	Sig                 []byte
}
```

#### Circuit

```go
func VerifyMintNftTx(
	api API, flag Variable,
	tx *MintNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints, nftBefore NftConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromMintNft(api, *tx)
	// verify params
	// check empty nft
	CheckEmptyNftNode(api, flag, nftBefore)
	// account index
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.ToAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// account name hash
	IsVariableEqual(api, flag, tx.ToAccountNameHash, accountsBefore[1].AccountNameHash)
	// content hash
	isZero := api.IsZero(tx.NftContentHash)
	IsVariableEqual(api, flag, isZero, 0)
	// gas asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	// should have enough balance
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	return pubData
}
```

### TransferNft

#### Description

This is a layer-2 transaction and is used for transferring NFTs to others in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 54                |

##### Structure

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

```go
func ConvertTxToTransferNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeTransferNft {
		logx.Errorf("[ConvertTxToMintNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToMintNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseTransferNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToMintNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type TransferNftTxInfo struct {
	FromAccountIndex  int64
	ToAccountIndex    int64
	ToAccountNameHash string
	NftIndex          int64
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	CallData          string
	CallDataHash      []byte
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyTransferNftTx(
	api API,
	flag Variable,
	tx *TransferNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromTransferNft(api, *tx)
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.ToAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// account name
	IsVariableEqual(api, flag, tx.ToAccountNameHash, accountsBefore[1].AccountNameHash)
	// asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	// nft info
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	IsVariableEqual(api, flag, tx.FromAccountIndex, nftBefore.OwnerAccountIndex)
	// should have enough balance
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	return pubData
}
```

### AtomicMatch

#### Description

This is a layer-2 transaction that will be used for buying or selling Nft in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 44                |

##### Structure

`Offer`:

| Name         | Size(byte) | Comment                                                      |
| ------------ | ---------- | ------------------------------------------------------------ |
| Type         | 1          | transaction type, 0 indicates this is a  `BuyNftOffer` , 1 indicate this is a  `SellNftOffer` |
| OfferId      | 3          | used to identify the oﬀer                                    |
| AccountIndex | 4          | who want to buy/sell nft                                     |
| AssetId      | 2          | the asset id which buyer/seller want to use pay for nft      |
| AssetAmount  | 5          | the asset amount                                             |
| ListedAt     | 8          | timestamp when the order is signed                           |
| ExpiredAt    | 8          | timestamp after which the order is invalid                   |
| Sig          | 64         | signature generated by buyer/seller_account_index's private key |

`AtomicMatch`(**below is the only info that will be uploaded on-chain**):

| Name                  | Size(byte) | Comment                    |
| --------------------- | ---------- | -------------------------- |
| TxType                | 1          | transaction type           |
| SubmitterAccountIndex | 4          | submitter account index    |
| BuyerAccountIndex     | 4          | buyer account index        |
| BuyerOfferId          | 3          | used to identify the offer |
| SellerAccountIndex    | 4          | seller account index       |
| SellerOfferId         | 3          | used to identify the offer |
| AssetId               | 2          | asset id                   |
| AssetAmount           | 5          | packed asset amount        |
| CreatorAmount         | 5          | packed creator amount      |
| TreasuryAmount        | 5          | packed treasury amount     |
| GasFeeAccountIndex    | 4          | gas fee account index      |
| GasFeeAssetId         | 2          | gas fee asset id           |
| GasFeeAssetAmount     | 2          | packed fee amount          |

```go
func ConvertTxToAtomicMatchPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeAtomicMatch {
		logx.Errorf("[ConvertTxToAtomicMatchPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToAtomicMatchPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToAtomicMatchPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.BuyOffer.AccountIndex)))
	buf.Write(Uint24ToBytes(txInfo.BuyOffer.OfferId))
	buf.Write(Uint32ToBytes(uint32(txInfo.SellOffer.AccountIndex)))
	buf.Write(Uint24ToBytes(txInfo.SellOffer.OfferId))
	buf.Write(Uint40ToBytes(txInfo.BuyOffer.NftIndex))
	buf.Write(Uint16ToBytes(uint16(txInfo.SellOffer.AssetId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	packedAmountBytes, err := AmountToPackedAmountBytes(txInfo.BuyOffer.AssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAmountBytes)
	creatorAmountBytes, err := AmountToPackedAmountBytes(txInfo.CreatorAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(creatorAmountBytes)
	treasuryAmountBytes, err := AmountToPackedAmountBytes(txInfo.TreasuryAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(treasuryAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type OfferTxInfo struct {
	Type         int64
	OfferId      int64
	AccountIndex int64
	NftIndex     int64
	AssetId      int64
	AssetAmount  *big.Int
	ListedAt     int64
	ExpiredAt    int64
	TreasuryRate int64
	Sig          []byte
}

type AtomicMatchTxInfo struct {
	AccountIndex      int64
	BuyOffer          *OfferTxInfo
	SellOffer         *OfferTxInfo
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	CreatorAmount     *big.Int
	TreasuryAmount    *big.Int
	Nonce             int64
	ExpiredAt         int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyAtomicMatchTx(
	api API, flag Variable,
	tx *AtomicMatchTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
	blockCreatedAt Variable,
	hFunc MiMC,
) (pubData [PubDataSizePerTx]Variable, err error) {
	pubData = CollectPubDataFromAtomicMatch(api, *tx)
	// verify params
	IsVariableEqual(api, flag, tx.BuyOffer.Type, 0)
	IsVariableEqual(api, flag, tx.SellOffer.Type, 1)
	IsVariableEqual(api, flag, tx.BuyOffer.AssetId, tx.SellOffer.AssetId)
	IsVariableEqual(api, flag, tx.BuyOffer.AssetAmount, tx.SellOffer.AssetAmount)
	IsVariableEqual(api, flag, tx.BuyOffer.NftIndex, tx.SellOffer.NftIndex)
	IsVariableEqual(api, flag, tx.BuyOffer.AssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.SellOffer.AssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.SellOffer.AssetId, accountsBefore[3].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[4].AccountIndex)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[4].AssetsInfo[1].AssetId)
	IsVariableLessOrEqual(api, flag, blockCreatedAt, tx.BuyOffer.ExpiredAt)
	IsVariableLessOrEqual(api, flag, blockCreatedAt, tx.SellOffer.ExpiredAt)
	IsVariableEqual(api, flag, nftBefore.NftIndex, tx.SellOffer.NftIndex)
	IsVariableEqual(api, flag, tx.BuyOffer.TreasuryRate, tx.SellOffer.TreasuryRate)
	// verify signature
	hFunc.Reset()
	buyOfferHash := ComputeHashFromOfferTx(tx.BuyOffer, hFunc)
	hFunc.Reset()
	notBuyer := api.IsZero(api.IsZero(api.Sub(tx.AccountIndex, tx.BuyOffer.AccountIndex)))
	notBuyer = api.And(flag, notBuyer)
	err = VerifyEddsaSig(notBuyer, api, hFunc, buyOfferHash, accountsBefore[1].AccountPk, tx.BuyOffer.Sig)
	if err != nil {
		return pubData, err
	}
	hFunc.Reset()
	sellOfferHash := ComputeHashFromOfferTx(tx.SellOffer, hFunc)
	hFunc.Reset()
	notSeller := api.IsZero(api.IsZero(api.Sub(tx.AccountIndex, tx.SellOffer.AccountIndex)))
	notSeller = api.And(flag, notSeller)
	err = VerifyEddsaSig(notSeller, api, hFunc, sellOfferHash, accountsBefore[2].AccountPk, tx.SellOffer.Sig)
	if err != nil {
		return pubData, err
	}
	// verify account index
	// submitter
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	// buyer
	IsVariableEqual(api, flag, tx.BuyOffer.AccountIndex, accountsBefore[1].AccountIndex)
	// seller
	IsVariableEqual(api, flag, tx.SellOffer.AccountIndex, accountsBefore[2].AccountIndex)
	// creator
	IsVariableEqual(api, flag, nftBefore.CreatorAccountIndex, accountsBefore[3].AccountIndex)
	// gas
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[4].AccountIndex)
	// verify buy offer id
	buyOfferIdBits := api.ToBinary(tx.BuyOffer.OfferId, 24)
	buyAssetId := api.FromBinary(buyOfferIdBits[7:]...)
	buyOfferIndex := api.Sub(tx.BuyOffer.OfferId, api.Mul(buyAssetId, OfferSizePerAsset))
	buyOfferIndexBits := api.ToBinary(accountsBefore[1].AssetsInfo[1].OfferCanceledOrFinalized, OfferSizePerAsset)
	for i := 0; i < OfferSizePerAsset; i++ {
		isZero := api.IsZero(api.Sub(buyOfferIndex, i))
		isCheckVar := api.And(isZero, flag)
		isCheck := api.Compiler().IsBoolean(isCheckVar)
		if isCheck {
			IsVariableEqual(api, 1, buyOfferIndexBits[i], 0)
		}
	}
	// verify sell offer id
	sellOfferIdBits := api.ToBinary(tx.SellOffer.OfferId, 24)
	sellAssetId := api.FromBinary(sellOfferIdBits[7:]...)
	sellOfferIndex := api.Sub(tx.SellOffer.OfferId, api.Mul(sellAssetId, OfferSizePerAsset))
	sellOfferIndexBits := api.ToBinary(accountsBefore[2].AssetsInfo[1].OfferCanceledOrFinalized, OfferSizePerAsset)
	for i := 0; i < OfferSizePerAsset; i++ {
		isZero := api.IsZero(api.Sub(sellOfferIndex, i))
		isCheckVar := api.And(isZero, flag)
		isCheck := api.Compiler().IsBoolean(isCheckVar)
		if isCheck {
			IsVariableEqual(api, 1, sellOfferIndexBits[i], 0)
		}
	}
	// buyer should have enough balance
	tx.BuyOffer.AssetAmount = UnpackAmount(api, tx.BuyOffer.AssetAmount)
	IsVariableLessOrEqual(api, flag, tx.BuyOffer.AssetAmount, accountsBefore[1].AssetsInfo[0].Balance)
	// submitter should have enough balance
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	return pubData, nil
}
```

### CancelOffer

#### Description

This is a layer-2 transaction and is used for canceling nft offer.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 16                |

##### Structure

| Name               | Size(byte) | Comment               |
| ------------------ | ---------- | --------------------- |
| TxType             | 1          | transaction type      |
| AccountIndex       | 4          | account index         |
| OfferId            | 3          | nft offer id          |
| GasFeeAccountIndex | 4          | gas fee account index |
| GasFeeAssetId      | 2          | gas fee asset id      |
| GasFeeAssetAmount  | 2          | packed fee amount     |

```go
func ConvertTxToCancelOfferPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeCancelOffer {
		logx.Errorf("[ConvertTxToCancelOfferPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToCancelOfferPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseCancelOfferTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToCancelOfferPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint24ToBytes(txInfo.OfferId))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type CancelOfferTxInfo struct {
	AccountIndex      int64
	OfferId           int64
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyCancelOfferTx(
	api API, flag Variable,
	tx *CancelOfferTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromCancelOffer(api, *tx)
	// verify params
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	offerIdBits := api.ToBinary(tx.OfferId, 24)
	assetId := api.FromBinary(offerIdBits[7:]...)
	IsVariableEqual(api, flag, assetId, accountsBefore[0].AssetsInfo[1].AssetId)
	// should have enough balance
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[1].Balance)
	return pubData
}
```

### WithdrawNft

#### Description

This is a layer-2 transaction and is used for withdrawing nft from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 162               |

##### Structure

| Name                   | Size(byte) | Comment                   |
| ---------------------- | ---------- | ------------------------- |
| TxType                 | 1          | transaction type          |
| AccountIndex           | 4          | account index             |
| CreatorAccountIndex    | 4          | creator account index     |
| CreatorTreasuryRate    | 2          | creator treasury rate     |
| NftIndex               | 5          | unique nft index          |
| CollectionId           | 2          | collection id             |
| NftL1Address           | 20         | nft layer-1 address       |
| ToAddress              | 20         | receiver address          |
| GasFeeAccountIndex     | 4          | gas fee account index     |
| GasFeeAssetId          | 2          | gas fee asset id          |
| GasFeeAssetAmount      | 2          | packed fee amount         |
| NftContentHash         | 32         | nft content hash          |
| NftL1TokenId           | 32         | nft layer-1 token id      |
| CreatorAccountNameHash | 32         | creator account name hash |

```go
func ConvertTxToWithdrawNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeWithdrawNft {
		logx.Errorf("[ConvertTxToWithdrawNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToWithdrawNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseWithdrawNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToWithdrawNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.ToAddress))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	chunk3 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(chunk3)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	return buf.Bytes(), nil
}
```

#### User transaction

```go
type WithdrawNftTxInfo struct {
	AccountIndex           int64
	CreatorAccountIndex    int64
	CreatorAccountNameHash []byte
	CreatorTreasuryRate    int64
	NftIndex               int64
	NftContentHash         []byte
	NftL1Address           string
	NftL1TokenId           *big.Int
	CollectionId           int64
	ToAddress              string
	GasAccountIndex        int64
	GasFeeAssetId          int64
	GasFeeAssetAmount      *big.Int
	ExpiredAt              int64
	Nonce                  int64
	Sig                    []byte
}
```

#### Circuit

```go
func VerifyWithdrawNftTx(
	api API,
	flag Variable,
	tx *WithdrawNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromWithdrawNft(api, *tx)
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// account name hash
	IsVariableEqual(api, flag, tx.CreatorAccountNameHash, accountsBefore[1].AccountNameHash)
	// collection id
	IsVariableEqual(api, flag, tx.CollectionId, nftBefore.CollectionId)
	// asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	// nft info
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, nftBefore.CreatorAccountIndex)
	IsVariableEqual(api, flag, tx.CreatorTreasuryRate, nftBefore.CreatorTreasuryRate)
	IsVariableEqual(api, flag, tx.AccountIndex, nftBefore.OwnerAccountIndex)
	IsVariableEqual(api, flag, tx.NftContentHash, nftBefore.NftContentHash)
	IsVariableEqual(api, flag, tx.NftL1TokenId, nftBefore.NftL1TokenId)
	IsVariableEqual(api, flag, tx.NftL1Address, nftBefore.NftL1Address)
	// have enough assets
	tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	return pubData
}
```

### FullExit

#### Description

This is a layer-1 transaction and is used for full exit assets from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 55                |

##### Structure

| Name            | Size(byte) | Comment            |
| --------------- | ---------- | ------------------ |
| TxType          | 1          | transaction type   |
| AccountIndex    | 4          | from account index |
| AssetId         | 2          | asset index        |
| AssetAmount     | 16         | state amount       |
| AccountNameHash | 32         | account name hash  |

```go
func ConvertTxToFullExitPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeFullExit {
		logx.Errorf("[ConvertTxToFullExitPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToFullExitPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseFullExitTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToFullExitPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	chunk := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(PrefixPaddingBufToChunkSize([]byte{}))
	return buf.Bytes(), nil
}
```

#### User transaction

| Name            | Size(byte) | Comment               |
| --------------- | ---------- | --------------------- |
| AccountNameHash | 32         | account name hash     |
| AssetAddress    | 20         | asset layer-1 address |

#### Circuit

```go
func VerifyFullExitTx(
	api API, flag Variable,
	tx FullExitTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromFullExit(api, tx)
	// verify params
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	return pubData
}
```

### FullExitNft

#### Description

This is a layer-1 transaction and is used for full exit NFTs from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
| ------ | ----------------- |
| 6      | 164               |

##### Structure

| Name                   | Size(byte) | Comment                   |
| ---------------------- | ---------- | ------------------------- |
| TxType                 | 1          | transaction type          |
| AccountIndex           | 4          | from account index        |
| CreatorAccountIndex    | 2          | creator account index     |
| CreatorTreasuryRate    | 2          | creator treasury rate     |
| NftIndex               | 5          | unique nft index          |
| CollectionId           | 2          | collection id             |
| NftL1Address           | 20         | nft layer-1 address       |
| AccountNameHash        | 32         | account name hash         |
| CreatorAccountNameHash | 32         | creator account name hash |
| NftContentHash         | 32         | nft content hash          |
| NftL1TokenId           | 32         | nft layer-1 token id      |

```go
func ConvertTxToFullExitNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	if oTx.TxType != commonTx.TxTypeFullExitNft {
		logx.Errorf("[ConvertTxToFullExitNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToFullExitNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseFullExitNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToFullExitNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk1 := SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	buf.Write(PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	return buf.Bytes(), nil
}
```

#### User transaction

| Name            | Size(byte) | Comment           |
| --------------- | ---------- | ----------------- |
| AccountNameHash | 32         | account name hash |
| NftIndex        | 5          | unique nft index  |

#### Circuit

```go
func VerifyFullExitNftTx(
	api API, flag Variable,
	tx FullExitNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints, nftBefore NftConstraints,
) (pubData [PubDataSizePerTx]Variable) {
	pubData = CollectPubDataFromFullExitNft(api, tx)
	// verify params
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	isCheck := api.IsZero(api.IsZero(tx.CreatorAccountNameHash))
	isCheck = api.And(flag, isCheck)
	IsVariableEqual(api, isCheck, tx.CreatorAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, isCheck, tx.CreatorAccountNameHash, accountsBefore[1].AccountNameHash)
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, nftBefore.CreatorAccountIndex)
	IsVariableEqual(api, flag, tx.CreatorTreasuryRate, nftBefore.CreatorTreasuryRate)
	isOwner := api.And(api.IsZero(api.Sub(tx.AccountIndex, nftBefore.OwnerAccountIndex)), flag)
	IsVariableEqual(api, isOwner, tx.NftContentHash, nftBefore.NftContentHash)
	IsVariableEqual(api, isOwner, tx.NftL1Address, nftBefore.NftL1Address)
	IsVariableEqual(api, isOwner, tx.NftL1TokenId, nftBefore.NftL1TokenId)
	tx.NftContentHash = api.Select(isOwner, tx.NftContentHash, 0)
	tx.NftL1Address = api.Select(isOwner, tx.NftL1Address, 0)
	tx.NftL1TokenId = api.Select(isOwner, tx.NftL1TokenId, 0)
	return pubData
}
```

## Smart contracts API

### Rollup contract

#### RegisterZNS

Register an ZNS account which is an ENS like domain for layer-1 and a short account name for your layer-2 account.

```js
function registerZNS(string calldata _name, address _owner, bytes32 _zkbnbPubKeyX, bytes32 _zkbnbPubKeyY) external payable nonReentrant
```

- `_name`: your favor account name
- `_owner`: account name layer-1 owner address
- `_zkbnbPubKeyX`: ZkBNB layer-2 public key X
- `_zkbnbPubKeyY`: ZkBNB layer-2 public key Y

#### CreatePair

Create a trading pair for layer-2.

```js
function createPair(address _tokenA, address _tokenB) external
```

- `_tokenA`: asset A address
- `_tokenB`: asset B address

#### UpdatePairRate

update a trading pair rate for layer-2:

```js
struct PairInfo {
    address tokenA;
    address tokenB;
    uint16 feeRate;
    uint32 treasuryAccountIndex;
    uint16 treasuryRate;
}

function updatePairRate(PairInfo memory _pairInfo) external
```

- `_assetAAddr`: asset A address
- `_assetBAddr`: asset B address
- `_feeRate`: fee rate
- `_treasuryAccountIndex`: the treasury account index in the layer-2 network
- `_treasuryRate`: treasury rate

#### Deposit BNB

Deposit BNB to Rollup - transfer BNB from user L1 address into Rollup account

```js
function depositBNB(bytes32 _accountNameHash) external payable
```

- `_accountNameHash`: The layer-2

#### Deposit BEP20

Deposit BEP20 assets to Rollup - transfer BEP20 assets from user L1 address into Rollup account

```js
function depositBEP20(
    IERC20 _token,
    uint104 _amount,
    bytes32 _accountNameHash
) external nonReentrant
```

- `_token`: valid BEP20 address
- `_amount`: deposit amount
- `_accountNameHash`: ZNS account name hash

#### Withdraw Pending BNB/BEP20

Withdraw BNB/BEP20 token to L1 - Transfer token from contract to owner

```js
function withdrawPendingBalance(
    address payable _owner,
    address _token,
    uint128 _amount
) external nonReentrant
```

- `_owner`: layer-1 address
- `_token`: asset address, `0` for BNB
- `_amount`: withdraw amount

#### Withdraw Pending Nft

Withdraw NFT to L1

```js
function withdrawPendingNFTBalance(uint40 _nftIndex) external
```

- `_nftIndex`: nft index

#### Censorship resistance

Register full exit request to withdraw all token balance from the account. The user needs to call it if she believes that her transactions are censored by the validator.

```js
function requestFullExit(bytes32 _accountNameHash, address _asset) public nonReentrant
```

- `_accountNameHash`: ZNS account name hash
- `_asset`: BEP20 asset address, `0` for BNB

Register full exit request to withdraw NFT tokens balance from the account. Users need to call it if they believe that their transactions are censored by the validator.

```js
function requestFullExitNFT(bytes32 _accountNameHash, uint32 _nftIndex) public nonReentrant
```

- `_accountNameHash`: ZNS account name hash
- `_nftIndex`: nft index

#### Desert mode

##### Withdraw funds

Withdraws token from Rollup to L1 in case of desert mode. User must provide proof that she owns funds.

// TODO

#### Rollup Operations

##### Commit block

Submit committed block data. Only active validator can make it. On-chain operations will be checked on contract and fulfilled on block verification.

```js
struct StoredBlockInfo {
    uint32 blockNumber;
    uint64 priorityOperations;
    bytes32 pendingOnchainOperationsHash;
    uint256 timestamp;
    bytes32 stateRoot;
    bytes32 commitment;
}

struct CommitBlockInfo {
    bytes32 newStateRoot;
    bytes publicData;
    uint256 timestamp;
    uint32[] publicDataOffsets;
    uint32 blockNumber;
}

function commitBlocks(
    StoredBlockInfo memory _lastCommittedBlockData,
    CommitBlockInfo[] memory _newBlocksData
)
external
```

`StoredBlockInfo`: block data that we store on BNB Smart Chain. We store hash of this structure in storage and pass it in tx arguments every time we need to access any of its field.

- `blockNumber`: rollup block number
- `priorityOperations`: priority operations count
- `pendingOnchainOperationsHash`: hash of all on-chain operations that have to be processed when block is finalized (verified)
- `timestamp`: block timestamp
- `stateRoot`: root hash of the rollup tree state
- `commitment`: rollup block commitment

`CommitBlockInfo`: data needed for new block commit

- `newStateRoot`: new layer-2 root hash
- `publicData`: public data of the executed rollup operations
- `timestamp`: block timestamp
- `publicDataOffsets`: list of on-chain operations offset
- `blockNumber`: rollup block number

`commitBlocks` and `commitOneBlock` are used for committing layer-2 transactions data on-chain.

- `_lastCommittedBlockData`: last committed block header
- `_newBlocksData`: pending commit blocks

##### Verify and execute blocks

Submit proofs of blocks and make it verified on-chain. Only active validator can make it. This block on-chain operations will be fulfilled.

```js
struct VerifyAndExecuteBlockInfo {
    StoredBlockInfo blockHeader;
    bytes[] pendingOnchainOpsPubData;
}

function verifyAndExecuteBlocks(VerifyAndExecuteBlockInfo[] memory _blocks, uint256[] memory _proofs) external
```

`VerifyAndExecuteBlockInfo`: block data that is used for verifying blocks

- `blockHeader`: related block header
- `pendingOnchainOpsPubdata`: public data of pending on-chain operations

`verifyBlocks`: is used for verifying block data and proofs

- `_blocks`: pending verify blocks
- `_proofs`: Groth16 proofs

#### Desert mode trigger

Checks if Desert mode must be entered. Desert mode must be entered in case of current BNB Smart Chain block number is higher than the oldest of existed priority requests expiration block number.

```js
function activateDesertMode() public returns (bool)
```

#### Revert blocks

Revert blocks that were not verified before deadline determined by `EXPECT_VERIFICATION_IN` constant.
```js
function revertBlocks(StoredBlockInfo[] memory _blocksToRevert) external
```

- `_blocksToRevert`: committed blocks to revert in reverse order starting from last committed.

#### Set default NFT factory

Set default NFT factory, which will be used for withdrawing NFT by default

```js
function setDefaultNFTFactory(NFTFactory _factory) external
```

- `_factory`: NFT factory address

#### Register NFT factory

Register NFT factory, which will be used for withdrawing NFT.

```js
function registerNFTFactory(
    string calldata _creatorAccountName,
    uint32 _collectionId,
    NFTFactory _factory
) external
```

- `_creatorAccountName`: NFT creator account name
- `_collectionId`: Collection id in the layer-2
- `_factory`: Address of NFTFactory

#### Get NFT factory for creator

Get NFT factory which will be used for withdrawing NFT for corresponding creator

```js
function getNFTFactory(bytes32 _creatorAccountNameHash, uint32 _collectionId) public view returns (address)
```

- `_creatorAccountNameHash`: Creator account name hash
- `_collectionId`: Collection id

### Governance contract

#### Change governor

Change current governor. The caller must be current governor.

```
function changeGovernor(address _newGovernor)
```

- `_newGovernor`: Address of the new governor

#### Add asset

Add asset to the list of networks assets. The caller must be current asset governance.

```js
function addAsset(address _asset) external
```

#### Set asset paused

Set asset status as paused or active. The caller must be current governor. It is impossible to create deposits of the paused assets.

```js
function setAssetPaused(address _assetAddress, bool _assetPaused) external
```

- `_assetAddress`: asset layer-1 address
- `_assetPausesd`: status

#### Set validator

Change validator status (active or not active). The caller must be current governor.

```js
function setValidator(address _validator, bool _active)
```

##### Change asset governance

```js
function changeAssetGovernance(AssetGovernance _newAssetGovernance) external
```

- `_newAssetGovernance`: New asset Governance

#### Check for governor

Validate that specified address is the token governance address

```js
function requireGovernor(address _address)
```

- `_address`: Address to check

#### Check for active validator

Validate that specified address is the active validator

```js
function requireActiveValidator(address _address)
```

- `_address`: Address to check

#### Check that asset address is valid

Validate asset address (it must be presented in assets list).

```js
function validateAssetAddress(address _assetAddr) external view returns (uint16)
```

- `_assetAddr`: Asset address

Returns: asset id.

### Asset Governance contract

#### Add asset

Collecting fees for adding an asset and passing the call to the `addAsset` function in the governance contract.

```js
function addAsset(address _assetAddress) external
```

- `_assetAddress`: BEP20 asset address

#### Set listing fee asset

Set new listing asset and fee, can be called only by governor.

```js
function setListingFeeAsset(IERC20 _newListingFeeAsset, uint256 _newListingFee) external
```

- `_newListingFeeAsset`: address of the asset in which fees will be collected
- `_newListingFee`: amount of tokens that will need to be paid for adding tokens

#### Set listing fee

Set new listing fee, can be called only by governor.

```js
function setListingFee(uint256 _newListingFee)
```

- `_newListingFee`: amount of assets that will need to be paid for adding tokens

#### Set lister

Enable or disable asset lister, if enabled new assets can be added by that address without payment, can be called only by governor.

```js
function setLister(address _listerAddress, bool _active)
```

- `_listerAddress`: address that can list tokens without fee
- `_active`: active flag

#### Set listing cap

Change maximum amount of assets that can be listed using this method, can be called only by governor.

```js
function setListingCap(uint16 _newListingCap)
```

- `_newListingCap`: max number of assets that can be listed using this contract

#### Set treasury

Change address that collects payments for listing assets, can be called only by governor.

```js
function setTreasury(address _newTreasury)
```

- `_newTreasury`: address that collects listing payments
