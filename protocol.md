

# ZecreyLegend Layer-2 Design

## Table of contents

## Glossary

- **L1**: layer-1 blockchain(BNB Chain)
- **Rollup**: layer-2 network (ZecreyLegend)
- **Owner**: a user who controls some assets in L2.
- **Operator**: entity operating the rollup.
- **Eventually**: happening within finite time.
- **Assets in rollup**: assets in L2 smart contract controlled by owners.
- **Rollup key**: owner's private key used to control deposited assets.
- **MiMC signature**: the result of signing the owner's message, using his private key, used in rollup internal transactions.

## Design

### Overview

ZecreyLegend implements a ZK rollup protocol (in short "rollup" below) for:

- BNB and BEP20 fungible token deposit and transfer
- AMM-based fungible token swap on L2
- BEP721 and BEP1155 non-fungible token deposit and transfer
- mint BEP721 or BEP1155 non-fungible tokens on L2
- NFT-marketplace on L2

General rollup workflow is as follows:

- Users can become owners in rollup by calling registerZNS in L1 to register a short name for L2;
- Owners can transfer assets to each other, mint NFT on L2 or make a swap on L2;
- Owners can withdraw assets under their control to an L1 address.

Rollup operation requires the assistance of an operator, who rolls transactions together, computes a zero-knowledge proof of the correct state transition, and affects the state transition by interacting with the rollup contract.

## Data format

### Data types

| Type           | Size(Byte) | Type     | Comment                                                      |
| -------------- | ---------- | -------- | ------------------------------------------------------------ |
| AccountIndex   | 4          | uint32   | Incremented number of accounts in Rollup. New account will have the next free id. Max value is 2^32 - 1 = 4.294967295 × 10^9 |
| AssetId        | 2          | uint16   | Incremented number of tokens in Rollup, max value is 65535   |
| PackedTxAmount | 5          | int64    | Packed transactions amounts are represented with 40 bit (5 byte) values, encoded as mantissa × 10^exponent where mantissa is represented with 35 bits, exponent is represented with 5 bits. This gives a range from 0 to 34359738368 × 10^31, providing 10 full decimal digit precision. |
| PackedFee      | 2          | uint16   | Packed fees must be represented with 2 bytes: 5 bit for exponent, 11 bit for mantissa. |
| StateAmount    | 16         | *big.Int | State amount is represented as uint128 with a range from 0 to ~3.4 × 10^38. It allows to represent up to 3.4 × 10^20 "units" if standard Ethereum's 18 decimal symbols are used. This should be a sufficient range. |
| Nonce          | 4          | uint32   | Nonce is the total number of executed transactions of the account. In order to apply the update of this state, it is necessary to indicate the current account nonce in the corresponding transaction, after which it will be automatically incremented. If you specify the wrong nonce, the changes will not occur. |
| EthAddress     | 20         | string   | To make an BNB Chain address from the BNB Chain's public key, all we need to do is to apply Keccak-256 hash function to the key and then take the last 20 bytes of the result. |
| Signature      | 64         | []byte   | Based on eddsa                                               |
| HashValue      | 32         | string   | hash value based on MiMC                                     |

### Amount packing

Mantissa and exponent parameters used in ZecreyLegend:

`amount = mantissa * radix^{exponent}`

| Type           | Exponent bit width | Mantissa bit width | Radix |
| -------------- | ------------------ | ------------------ | ----- |
| PackedTxAmount | 5                  | 35                 | 10    |
| PackedFee      | 5                  | 11                 | 10    |

### State Merkle Tree(height)

We have 3 unique trees: `AccountTree(32)`, `LiquidityTree(16)`, `NftTree(40)` and one sub tree `AssetTree(16)` which belongs to `AccountTree(32)`. The empty leaf for all of the trees is just set every attribute as `0` for every node.

#### AccountTree

`AccountTree` is used for storing all accounts info and the node of the account tree is:

```go
type AccountNode struct{
    AccountNameHash string // bytes32
    PubKey string // bytes32
    Nonce int64
    AssetRoot string // bytes32
}
```

Leaf hash computation:

```go
func ComputeAccountLeafHash(
	accountNameHash string, 
    pk string, 
    nonce int64,
	assetRoot []byte,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	buf.Write(common.FromHex(accountNameHash))
	err = util.WritePkIntoBuf(&buf, pk)
	util.WriteInt64IntoBuf(&buf, nonce)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] unable to write pk into buf: %s", err.Error())
		return nil, err
	}
	buf.Write(assetRoot)
	hFunc.Reset()
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}
```

##### AssetTree

`AssetTree` is a sub tree of `AccountTree` and it stores all of the assets `balance` and `lpAmount`. The node of asset tree is:

```go
type AssetNode struct {
	Balance  string
	LpAmount string
}
```

Leaf hash computation:

```go
func ComputeAccountAssetLeafHash(
	balance string,
	lpAmount string,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	err = util.WriteStringBigIntIntoBuf(&buf, balance)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	err = util.WriteStringBigIntIntoBuf(&buf, lpAmount)
	if err != nil {
		logx.Errorf("[ComputeAccountAssetLeafHash] invalid balance: %s", err.Error())
		return nil, err
	}
	hFunc.Write(buf.Bytes())
	return hFunc.Sum(nil), nil
}
```

#### LiquidityTree

`LiquidityTree` is used for storing all of the liquidities info and the node of the liquidity tree is:

```go
type LiquidityNode struct {
    AssetAId      int64
    AssetA        string
    AssetBId      int64
    AssetB        string
    LpAmount      string
}
```

The liquidity pair is first initialized by `CreatePair` tx and will be changed by `Swap`, `AddLiquidity` and `RemoveLiquidity` txs.

Leaf hash computation:

```go
func ComputeLiquidityAssetLeafHash(
	assetAId int64,
	assetA string,
	assetBId int64,
	assetB string,
    lpAmount string,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	util.WriteInt64IntoBuf(&buf, assetAId)
	err = util.WriteStringBigIntIntoBuf(&buf, assetA)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	util.WriteInt64IntoBuf(&buf, assetBId)
	err = util.WriteStringBigIntIntoBuf(&buf, assetB)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
    err = util.WriteStringBigIntIntoBuf(&buf, lpAmount)
	if err != nil {
		logx.Errorf("[ComputeLiquidityAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}
```

#### NftTree

`NftTree` is used for storing all of the nfts and the node info is:

```go
type NftNode struct {
    CreatorAccountIndex int64
    OwnerAccountIndex   int64
    AssetId             int64
    AssetAmount         string
    NftContentHash      string
    NftL1TokenId        string
    NftL1Address        string
    CreatorTreasuryRate int64
}
```

Leaf hash computation:

```go
func ComputeNftAssetLeafHash(
	creatorIndex int64,
	nftContentHash string,
	assetId int64,
	assetAmount string,
	nftL1Address string,
	nftL1TokenId string,
    creatorTreasuryRate int64,
) (hashVal []byte, err error) {
	hFunc := mimc.NewMiMC()
	var buf bytes.Buffer
	util.WriteInt64IntoBuf(&buf, creatorIndex)
	buf.Write(common.FromHex(nftContentHash))
	util.WriteInt64IntoBuf(&buf, assetId)
	err = util.WriteStringBigIntIntoBuf(&buf, assetAmount)
	if err != nil {
		logx.Errorf("[ComputeNftAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
	err = util.WriteAddressIntoBuf(&buf, nftL1Address)
	if err != nil {
		logx.Errorf("[ComputeNftAssetLeafHash] unable to write address to buf: %s", err.Error())
		return nil, err
	}
	err = util.WriteStringBigIntIntoBuf(&buf, nftL1TokenId)
	if err != nil {
		logx.Errorf("[ComputeNftAssetLeafHash] unable to write big int to buf: %s", err.Error())
		return nil, err
	}
    util.WriteInt64IntoBuf(&buf, creatorTreasuryRate)
	hFunc.Write(buf.Bytes())
	hashVal = hFunc.Sum(nil)
	return hashVal, nil
}
```

#### StateRoot

`StateRoot` is the final root that shows the final layer-2 state and will be stored on L1. It is computed by the root of `AccountTree`, `LiquidityTree` and `NftTree`. The computation of `StateRoot` works as follows:

`StateRoot = MiMC(AccountRoot || LiquidityRoot || NftRoot)`

## ZecreyLegend Transactions

ZecreyLegend transactions are divided into Rollup transactions (initiated inside Rollup by a Rollup account) and Priority operations (initiated on the mainchain by an BNB Chain account).

Rollup transactions:

- EmptyTx
- Transfer
- Swap
- AddLiquidity
- RemoveLiquidity
- Withdraw
- MintNft
- TransferNft
- SetNftPrice
- BuyNft
- WithdrawNft

Priority operations:

- RegisterZNS
- CreatePair
- Deposit
- DepositNft
- FullExit
- FullExitNft

### Rollup transaction lifecycle

1. User creates a `Transaction` or a `Priority operation`.
2. After processing this request, operator creates a `Rollup operation` and adds it to the block.
3. Once the block is complete, operator submits it to the ZecreyLegend smart contract as a block commitment. Part of the logic of some `Rollup transaction` is checked by the smart contract.
4. The proof for the block is submitted to the ZecreyLegend smart contract as the block verification. If the verification succeeds, the new state is considered finalized.

### EmptyTx

#### Description

No effects.

#### Onchain operation

##### Size

1 byte

##### Structure

| Field  | Size(byte) | Value/type | Description      |
| ------ | ---------- | ---------- | ---------------- |
| TxType | 1          | `0x00`     | Transaction type |

#### User transaction

No user transaction

### RegisterZNS

#### Description

This is a layer-1 transaction and a user needs to call this method first to register a layer-2 account.

#### Onchain operation

##### Size

101 byte

##### Structure

| Name            | Size(byte) | Comment                        |
| --------------- | ---------- | ------------------------------ |
| TxType          | 1          | transaction type               |
| AccountIndex    | 4          | unique account index           |
| AccountName     | 32         | account name                   |
| AccountNameHash | 32         | hash value of the account name |
| PubKey          | 32         | layer-2 account's public key   |

#### User transaction

| Name        | Size(byte) | Comment                      |
| ----------- | ---------- | ---------------------------- |
| AccountName | 32         | account name                 |
| Owner       | 20         | account layer-1 address      |
| PubKey      | 32         | layer-2 account's public key |

#### Circuit

```go
func VerifyRegisterZNSTx(api API, flag Variable, accountsBefore [NbAccountsPerTx]AccountConstraints) {
	CheckEmptyAccountNode(api, flag, accountsBefore[0])
}
```

### CreatePair

#### Description

This is a layer-1 transaction and is used for creating a trading pair for L2.

#### Onchain operation

##### Size

5 byte

##### Structure

| Name     | Size(byte) | Comment            |
| -------- | ---------- | ------------------ |
| TxType   | 1          | transaction type   |
| AssetAId | 2          | unique asset index |
| AssetBId | 2          | unique asset index |

#### User transaction

| Name          | Size(byte) | Comment                 |
| ------------- | ---------- | ----------------------- |
| AssetAAddress | 20         | asset a layer-1 address |
| AssetBAddress | 20         | asset b layer-1 address |

#### Circuit

```go
func VerifyCreatePairTx(api API, flag Variable, tx CreatePairTxConstraints, liquidityBefore LiquidityConstraints) {
	// verify params
	IsVariableEqual(api, flag, tx.PairIndex, liquidityBefore.PairIndex)
	CheckEmptyLiquidityNode(api, flag, liquidityBefore)
}
```

### Deposit

#### Description

This is a layer-1 transaction and is used for depositing assets into the layer-2 account.

#### Onchain operation

##### Size

55 byte

##### Structure

| Name            | Size(byte) | Comment           |
| --------------- | ---------- | ----------------- |
| TxType          | 1          | transaction type  |
| AccountIndex    | 4          | account index     |
| AccountNameHash | 32         | account name hash |
| AssetId         | 2          | asset index       |
| AssetAmount     | 16         | state amount      |

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
func VerifyDepositTx(api API, flag Variable, tx DepositTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints) {
	// verify params
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
}
```

### DepositNft

#### Description

This is a layer-1 transaction and is used for depositing nfts into the layer-2 account.

#### Onchain operation

##### Size

94 byte

##### Structure

| Name                | Size(byte) | Comment               |
| ------------------- | ---------- | --------------------- |
| TxType              | 1          | transaction type      |
| AccountIndex        | 4          | account index         |
| NftIndex            | 5          | unique index of a nft |
| NftContentHash      | 32         | nft content hash      |
| NftL1Address        | 20         | nft layer-1 address   |
| NftL1TokenId        | 32         | nft layer-1 token id  |
| CreatorTreasuryRate | 2          | creator treasury rate |

#### User transaction

| Name                | Size(byte) | Comment                      |
| ------------------- | ---------- | ---------------------------- |
| AccountNameHash     | 32         | account name hash            |
| AssetAddress        | 20         | nft contract layer-1 address |
| NftTokenId          | 32         | nft layer-1 token id         |
| CreatorTreasuryRate | 2          | creator treasury rate        |

#### Circuit

```go
func VerifyDepositNftTx(
	api API,
	flag Variable,
	tx DepositNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) {
	// verify params
	// check empty nft
	CheckEmptyNftNode(api, flag, nftBefore)
	// account index
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	// account name hash
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
}
```

### Transfer

#### Description

This is a layer-2 transaction and is used for transfering assets in the layer-2 network.

#### Onchain operation

##### Size

56 byte

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

#### User transaction

```go
type TransferTxInfo struct {
	FromAccountIndex  int64
	ToAccountIndex    int64
	ToAccountName     string
	AssetId           int64
	AssetAmount       *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	Memo              string
	CallData          string
	CallDataHash      []byte
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyTransferTx(api API, flag Variable, tx TransferTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.ToAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// gas asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[1].AssetId)
	// should have enough balance
	IsVariableLessOrEqual(api, flag, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[1].Balance)
}
```

### Swap

#### Description

This is a layer-2 transaction and is used for making a swap for assets in the layer-2 network.

#### Onchain operation

##### Size

31 byte

##### Structure

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

#### User transaction

```go
type SwapTxInfo struct {
	FromAccountIndex       int64
	PairIndex              int64
	AssetAId               int64
	AssetAAmount           *big.Int
	AssetBId               int64
	AssetBMinAmount        *big.Int
	AssetBAmountDelta      *big.Int
	PoolAAmount            *big.Int
	PoolBAmount            *big.Int
	FeeRate                int64 // 0.3 * 10000
	TreasuryAccountIndex   int64
	TreasuryRate           int64
	TreasuryFeeAmountDelta *big.Int
	GasAccountIndex        int64
	GasFeeAssetId          int64
	GasFeeAssetAmount      *big.Int
	Nonce                  int64
	Sig                    []byte
}
```

#### Circuit

```go
func VerifySwapTx(api API, flag Variable, tx SwapTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints, liquidityBefore LiquidityConstraints) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.TreasuryAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// pair index
	IsVariableEqual(api, flag, tx.PairIndex, liquidityBefore.PairIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetAId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetBId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.AssetAId, accountsBefore[1].AssetsInfo[0].AssetId)
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
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	// should have enough assets
	IsVariableLessOrEqual(api, flag, tx.AssetBMinAmount, tx.AssetBAmountDelta)
	IsVariableLessOrEqual(api, flag, tx.AssetAAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[2].Balance)
	// pool info
	isSameAsset = api.And(flag, isSameAsset)
	isDifferentAsset = api.And(flag, isSameAsset)
	IsVariableLessOrEqual(api, isSameAsset, tx.PoolAAmount, liquidityBefore.AssetA)
	IsVariableLessOrEqual(api, isSameAsset, tx.PoolBAmount, liquidityBefore.AssetB)
	IsVariableLessOrEqual(api, isDifferentAsset, tx.PoolAAmount, liquidityBefore.AssetB)
	IsVariableLessOrEqual(api, isDifferentAsset, tx.PoolBAmount, liquidityBefore.AssetA)
	// verify AMM
	k := api.Mul(tx.PoolAAmount, tx.PoolBAmount)
	// check treasury fee amount
	treasuryAmount := api.Div(api.Mul(tx.AssetAAmount, tx.TreasuryRate), 10000)
	IsVariableEqual(api, flag, tx.TreasuryFeeAmountDelta, treasuryAmount)
	poolADelta := api.Sub(tx.AssetAAmount, tx.TreasuryFeeAmountDelta)
	kPrime := api.Mul(api.Add(tx.PoolAAmount, poolADelta), api.Sub(tx.PoolBAmount, tx.AssetBAmountDelta))
	api.AssertIsLessOrEqual(k, kPrime)
}
```

### AddLiquidity

#### Description

This is a layer-2 transaction and is used for adding liquidity for a trading pair in the layer-2 network.

#### Onchain operation

##### Size

30 byte

##### Structure

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
	PoolAAmount       *big.Int
	PoolBAmount       *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyAddLiquidityTx(api API, flag Variable, tx AddLiquidityTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints, liquidityBefore LiquidityConstraints) {
	// check params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetAId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetBId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.AssetAId, liquidityBefore.AssetAId)
	IsVariableEqual(api, flag, tx.AssetBId, liquidityBefore.AssetBId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[2].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// check if the user has enough balance
	IsVariableLessOrEqual(api, flag, tx.AssetAAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.AssetBAmount, accountsBefore[0].AssetsInfo[1].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[2].Balance)
	IsVariableEqual(api, flag, tx.PoolAAmount, liquidityBefore.AssetA)
	IsVariableEqual(api, flag, tx.PoolBAmount, liquidityBefore.AssetB)
	// verify LP
	Delta_LPCheck := api.Mul(tx.AssetAAmount, tx.AssetBAmount)
	LPCheck := api.Mul(tx.LpAmount, tx.LpAmount)
	api.AssertIsLessOrEqual(LPCheck, Delta_LPCheck)
	// TODO verify AMM info
	l := api.Mul(tx.PoolBAmount, tx.AssetAAmount)
	r := api.Mul(tx.PoolAAmount, tx.AssetBAmount)
	api.AssertIsEqual(l, r)
}
```

### RemoveLiquidity

#### Description

This is a layer-2 transaction and is used for removing liquidity for a trading pair in the layer-2 network.

#### Onchain operation

##### Size

30 byte

##### Structure

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
	PoolAAmount       *big.Int
	PoolBAmount       *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyRemoveLiquidityTx(api API, flag Variable, tx RemoveLiquidityTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints, liquidityBefore LiquidityConstraints) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetAId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetBId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[3].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetAId, liquidityBefore.AssetAId)
	IsVariableEqual(api, flag, tx.AssetBId, liquidityBefore.AssetBId)
	// should have enough lp
	IsVariableLessOrEqual(api, flag, tx.LpAmount, accountsBefore[0].AssetsInfo[2].LpAmount)
	// enough balance
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	// verify LP
	Delta_LPCheck := api.Mul(tx.AssetAAmountDelta, tx.AssetBAmountDelta)
	LPCheck := api.Mul(tx.LpAmount, tx.LpAmount)
	IsVariableLessOrEqual(api, flag, Delta_LPCheck, LPCheck)
	IsVariableLessOrEqual(api, flag, tx.AssetAMinAmount, tx.AssetAAmountDelta)
	IsVariableLessOrEqual(api, flag, tx.AssetBMinAmount, tx.AssetBAmountDelta)
}
```

### Withdraw

#### Description

This is a layer-2 transaction and is used for withdrawing assets from the layer-2 to the layer-1.

#### Onchain operation

##### Size

51 byte

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
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyWithdrawTx(api API, flag Variable, tx WithdrawTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// should have enough assets
	IsVariableLessOrEqual(api, flag, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[1].Balance)
}
```

### MintNft

#### Description

This is a layer-2 transaction and is used for minting nfts in the layer-2 network.

#### Onchain operation

##### Size

54 byte

##### Structure

| Name                | Size(byte) | Comment                |
| ------------------- | ---------- | ---------------------- |
| TxType              | 1          | transaction type       |
| FromAccountIndex    | 4          | from account index     |
| ToAccountIndex      | 4          | receiver account index |
| NftIndex            | 5          | unique nft index       |
| NftContentHash      | 32         | nft content hash       |
| GasFeeAccountIndex  | 4          | gas fee account index  |
| GasFeeAssetId       | 2          | gas fee asset id       |
| GasFeeAssetAmount   | 2          | packed fee amount      |
| CreatorTreasuryRate | 2          | creator treasury rate  |

#### User transaction

```go
type MintNftTxInfo struct {
	CreatorAccountIndex int64
	ToAccountIndex      int64
	ToAccountName       string
	NftIndex            int64
	NftContentHash      string
	NftName             string
	NftIntroduction     string
	NftAttributes       string
	NftCollectionId     int64
	CreatorFeeRate      int64
	GasAccountIndex     int64
	GasFeeAssetId       int64
	GasFeeAssetAmount   *big.Int
	Nonce               int64
	Sig                 []byte
}
```

#### Circuit

```go
func VerifyMintNftTx(api API, flag Variable, tx MintNftTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints, nftBefore NftConstraints) {
	// verify params
	// check empty nft
	CheckEmptyNftNode(api, flag, nftBefore)
	// account index
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.ToAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[2].AccountIndex)
	// gas asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	// should have enough balance
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
}
```

### TransferNft

#### Description

This is a layer-2 transaction and is used for transfering nfts to others in the layer-2 network.

#### Onchain operation

##### Size

54 byte

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

#### User transaction

```go
type TransferNftTxInfo struct {
	FromAccountIndex  int64
	ToAccountIndex    int64
	ToAccountName     string
	NftIndex          int64
	NftContentHash    string
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	CallData          string
	CallDataHash      []byte
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyTransferNftTx(
	api API,
	flag Variable,
	tx TransferNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// nft info
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	IsVariableEqual(api, flag, tx.FromAccountIndex, nftBefore.OwnerAccountIndex)
	// should have enough balance
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
}
```

### SetNftPrice

#### Description

This is a layer-2 transaction and is used for setting nft price in the layer-2 network.

#### Onchain operation

##### Size

25 byte

##### Structure

| Name               | Size(byte) | Comment               |
| ------------------ | ---------- | --------------------- |
| TxType             | 1          | transaction type      |
| FromAccountIndex   | 4          | from account index    |
| NftIndex           | 5          | unique nft index      |
| AssetId            | 2          | asset index           |
| AssetAmount        | 5          | packed amount         |
| GasFeeAccountIndex | 4          | gas fee account index |
| GasFeeAssetId      | 2          | gas fee asset id      |
| GasFeeAssetAmount  | 2          | packed fee amount     |

#### User transaction

```go
type SetNftPriceTxInfo struct {
	AccountIndex      int64
	NftIndex          int64
	NftContentHash    string
	AssetId           int64
	AssetAmount       *big.Int
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifySetNftPriceTx(api API, flag Variable, tx SetNftPriceTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints, nftBefore NftConstraints) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// nft info
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	IsVariableEqual(api, flag, tx.AccountIndex, nftBefore.OwnerAccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	// should have enough assets
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
}
```

### BuyNft

#### Description

This is a layer-2 transaction and is used for buying nfts in the layer-2 network.

#### Onchain operation

##### Size

37 byte

##### Structure

| Name                    | Size(byte) | Comment                |
| ----------------------- | ---------- | ---------------------- |
| TxType                  | 1          | transaction type       |
| BuyerAccountIndex       | 4          | buyer account index    |
| OwnerAccountIndex       | 4          | owner account index    |
| NftIndex                | 5          | unique nft index       |
| AssetId                 | 2          | asset index            |
| AssetAmount             | 5          | packed amount          |
| GasFeeAccountIndex      | 4          | gas fee account index  |
| GasFeeAssetId           | 2          | gas fee asset id       |
| GasFeeAssetAmount       | 2          | packed fee amount      |
| TreasuryFeeAccountIndex | 4          | treasury account index |
| TreasuryFeeAmount       | 2          | packed fee             |
| CreatorFeeAmount        | 2          | packed fee             |

#### User transaction

```go
type BuyNftTxInfo struct {
	BuyerAccountIndex    int64
	OwnerAccountIndex    int64
	NftIndex             int64
	NftContentHash       string
	AssetId              int64
	AssetAmount          *big.Int
	TreasuryFeeRate      int64
	TreasuryAccountIndex int64
	CreatorAccountIndex  int64
	CreatorFeeRate       int64
	GasAccountIndex      int64
	GasFeeAssetId        int64
	GasFeeAssetAmount    *big.Int
	Nonce                int64
	Sig                  []byte
}
```

#### Circuit

```go
func VerifyBuyNftTx(api API, flag Variable, tx BuyNftTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints, nftBefore NftConstraints) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.BuyerAccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.OwnerAccountIndex, accountsBefore[1].AccountIndex)
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[2].AccountIndex)
	IsVariableEqual(api, flag, tx.TreasuryAccountIndex, accountsBefore[3].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[4].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[2].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[3].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[1].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[4].AssetsInfo[0].AssetId)
	// nft info
	IsVariableEqual(api, flag, tx.CreatorAccountIndex, nftBefore.CreatorAccountIndex)
	IsVariableEqual(api, flag, tx.OwnerAccountIndex, nftBefore.OwnerAccountIndex)
	IsVariableEqual(api, flag, tx.AssetId, nftBefore.AssetId)
	IsVariableEqual(api, flag, tx.AssetAmount, nftBefore.AssetAmount)
	IsVariableEqual(api, flag, tx.CreatorTreasuryRate, nftBefore.CreatorTreasuryRate)
	// TODO treasury amount check
	// should have enough assets
	IsVariableLessOrEqual(api, flag, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[1].Balance)
}
```

### WithdrawNft

#### Description

This is a layer-2 transaction and is used for withdrawing nft from the layer-2 to the layer-1.

#### Onchain operation

##### Size

102 byte

##### Structure

| Name               | Size(byte) | Comment               |
| ------------------ | ---------- | --------------------- |
| TxType             | 1          | transaction type      |
| AccountIndex       | 4          | account index         |
| NftIndex           | 5          | unique nft index      |
| NftContentHash     | 32         | nft content hash      |
| NftL1Address       | 20         | nft layer-1 address   |
| NftL1TokenId       | 32         | nft layer-1 token id  |
| GasFeeAccountIndex | 4          | gas fee account index |
| GasFeeAssetId      | 2          | gas fee asset id      |
| GasFeeAssetAmount  | 2          | packed fee amount     |

#### User transaction

```go
type WithdrawNftTxInfo struct {
	AccountIndex int64
	NftIndex          int64
	NftContentHash    string
	NftL1Address      string
	NftL1TokenId      *big.Int
	ToAddress         string
	ProxyAddress      string
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	Nonce             int64
	Sig               []byte
}
```

#### Circuit

```go
func VerifyWithdrawNftTx(
	api API,
	flag Variable,
	tx WithdrawNftTxConstraints,
	accountsBefore [NbAccountsPerTx]AccountConstraints,
	nftBefore NftConstraints,
) {
	// verify params
	// account index
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.GasAccountIndex, accountsBefore[1].AccountIndex)
	// asset id
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[1].AssetsInfo[0].AssetId)
	// nft info
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	IsVariableEqual(api, flag, tx.AccountIndex, nftBefore.OwnerAccountIndex)
	// have enough assets
	IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
}
```

### FullExit

#### Description

This is a layer-1 transaction and is used for full exit assets from the layer-2 to the layer-1.

#### Onchain operation

##### Size

55 byte

##### Structure

| Name            | Size(byte) | Comment            |
| --------------- | ---------- | ------------------ |
| TxType          | 1          | transaction type   |
| AccountIndex    | 4          | from account index |
| AccountNameHash | 32         | account name hash  |
| AssetId         | 2          | asset index        |
| AssetAmount     | 16         | state amount       |

#### User transaction

| Name            | Size(byte) | Comment               |
| --------------- | ---------- | --------------------- |
| AccountNameHash | 32         | account name hash     |
| AssetAddress    | 20         | asset layer-1 address |

#### Circuit

```go
func VerifyFullExitTx(api API, flag Variable, tx FullExitTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints) {
	// verify params
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
	IsVariableEqual(api, flag, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
}
```

### FullExitNft

#### Description

This is a layer-1 transaction and is used for full exit nfts from the layer-2 to the layer-1.

#### Onchain operation

##### Size

94 byte

##### Structure

| Name           | Size(byte) | Comment              |
| -------------- | ---------- | -------------------- |
| TxType         | 1          | transaction type     |
| AccountIndex   | 4          | from account index   |
| NftIndex       | 5          | unique nft index     |
| NftContentHash | 32         | nft content hash     |
| NftL1Address   | 20         | nft layer-1 address  |
| NftL1TokenId   | 32         | nft layer-1 token id |

#### User transaction

| Name            | Size(byte) | Comment           |
| --------------- | ---------- | ----------------- |
| AccountNameHash | 32         | account name hash |
| NftIndex        | 5          | unique nft index  |

#### Circuit

```go
func VerifyFullExitNftTx(api API, flag Variable, tx FullExitNftTxConstraints, accountsBefore [NbAccountsPerTx]AccountConstraints, nftBefore NftConstraints) {
	// verify params
	IsVariableEqual(api, flag, tx.AccountNameHash, accountsBefore[0].AccountNameHash)
	IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
	IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
	isOwner := api.And(api.IsZero(api.Sub(tx.AccountIndex, nftBefore.OwnerAccountIndex)), flag)
	IsVariableEqual(api, isOwner, tx.NftContentHash, nftBefore.NftContentHash)
	IsVariableEqual(api, isOwner, tx.NftL1Address, nftBefore.NftL1Address)
	IsVariableEqual(api, isOwner, tx.NftL1TokenId, nftBefore.NftL1TokenId)
	tx.NftContentHash = api.Select(isOwner, tx.NftContentHash, 0)
	tx.NftL1Address = api.Select(isOwner, tx.NftL1Address, 0)
	tx.NftL1TokenId = api.Select(isOwner, tx.NftL1TokenId, 0)
}
```

## Smart contracts API

### Rollup contract

#### RegisterZNS

Register an ZNS account which is an ENS like domain for layer-1 and a short account name for your layer-2 account.

```js
function registerZNS(string calldata _name, address _owner, bytes32 _zecreyPubKey) external nonReentrant
```

- `_name`: your favor account name
- `_owner`: account name layer-1 owner address
- `_zecreyPubKey`: zecrey layer-2 public key

#### CreatePair

Create a trading pair for layer-2.

```js
function createPair(address _assetAAddr, address _assetBAddr) external nonReentrant
```

- `_assetAAddr`: asset A address
- `_assetBAddr`: asset B address

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

// TODO

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

Submit committed block data. Only active validator can make it. Onchain operations will be checked on contract and fulfilled on block verification.

```js
struct BlockHeader {
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
    BlockHeader memory _lastCommittedBlockData,
    CommitBlockInfo[] memory _newBlocksData
)
external
nonReentrant
```

`BlockHeader`: block data that we store on BNB Chain. We store hash of this structure in storage and pass it in tx arguments every time we need to access any of its field.

- `blockNumber`: rollup block number
- `priorityOperations`: priority operations count
- `pendingOnchainOperationsHash`: hash of all onchain operations that have to be processed when block is finalized (verified)
- `timestamp`: block timestamp
- `stateRoot`: root hash of the rollup tree state
- `commitment`: rollup block commitment

`CommitBlockInfo`: data needed for new block commit

- `newStateRoot`: new layer-2 root hash
- `publicData`: public data of the executed rollup operations
- `timestamp`: block timestamp
- `publicDataOffsets`: list of onchain operations offset
- `blockNumber`: rollup block number

`commitBlocks` and `commitOneBlock` are used for committing layer-2 transactions data onchain.

- `_lastCommittedBlockData`: last committed block header
- `_newBlocksData`: pending commit blocks

##### Verify blocks

Submit proofs of blocks and make it verified onchain. Only active validator can make it. This block onchain operations will be fulfilled.

```js
struct VerifyBlockInfo {
    BlockHeader blockHeader;
    bytes[] pendingOnchainOpsPubdata;
}

function verifyBlocks(VerifyBlockInfo[] memory _blocks, uint256[] memory _proofs) external nonReentrant
```

`VerifyBlockInfo`: block data that is used for verifying blocks

- `blockHeader`: related block header
- `pendingOnchainOpsPubdata`: public data of pending onchain operations

`verifyBlocks`: is used for verifying block data and proofs

- `_blocks`: pending verify blocks
- `_proofs`: Groth16 proofs

#### Desert mode trigger

Checks if Desert mode must be entered. Desert mode must be entered in case of current BNB Chain block number is higher than the oldest of existed priority requests expiration block number.

```js
function activateDesertMode() public returns (bool)
```

#### Revert blocks

// TODO

Revert blocks that were not verified before deadline determined by `EXPECT_VERIFICATION_IN` constant. The caller must be valid operator.

```js
function revertBlocks(BlockHeader[] memory _blocksToRevert) external
```

- `_blocksToRevert`: committed blocks to revert in reverse order starting from last committed.

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

Set asset status as paused or actived. The caller must be current governor. Its impossible to create deposits of the paused assets.

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

#### Set default NFT factory

// TODO

Register factory, which will be use for withdrawing NFT by default

```js
function setDefaultNFTFactory(address _factory)
```

- `_factory`: NFT factory address

#### Get NFT factory for creator

// TODO

Get NFT factory which will be used for withdrawing NFT for corresponding creator

```js
function getNFTFactory(uint32 _creatorAccountId, address _creatorAddress)
```

- `_creatorAccountId`: Creator account id
- `_creatorAddress`: Creator address

### Asset Governance contract

#### Add asset

Collects fees for adding a asset and passes the call to the `addAsset` function in the governance contract.

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
