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
- BEP721 non-fungible token deposit and transfer
- mint BEP721 non-fungible tokens on L2
- NFT-marketplace on L2

General rollup workflow is as follows:

- Users can become owners in rollup by calling registerZNS in L1 to register a short name for L2;
- Owners can transfer assets to each other, mint NFT on L2;
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

We have 3 unique trees: `AccountTree(32)`, `NftTree(40)` and one sub-tree `AssetTree(16)` which 
belongs to `AccountTree(32)`. The empty leaf for all the trees is just set every attribute as `0` for every node.

#### AccountTree

`AccountTree` is used for storing all accounts info and the node of the account tree is:

```go
type AccountNode struct{
    L1Address string // bytes20
    PubKey string // bytes32
    Nonce int64
    CollectionNonce int64
    AssetRoot string // bytes32
}
```

Leaf hash computation:

```go
func ComputeAccountLeafHash(
   l1Address string,
   pk string,
   nonce int64,
   collectionNonce int64,
   assetRoot []byte,
   ctx context.Context,
) (hashVal []byte, err error) {
   var e0 *fr.Element
   if l1Address == "" {
   e0 = &fr.Element{0, 0, 0, 0}
   e0.SetBytes([]byte{})
   } else {
   e0, err = txtypes.FromBytesToFr(common.FromHex(l1Address))
   if err != nil {
   return nil, err
   }
   }
   pubKey, err := common2.ParsePubKey(pk)
   if err != nil {
   return nil, err
   }
   e1 := &pubKey.A.X
   e2 := &pubKey.A.Y
   e3 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(nonce))
   e4 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(collectionNonce))
   e5 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(assetRoot))
   ele := GMimcElements([]*fr.Element{e0, e1, e2, e3, e4, e5})
   hash := ele.Bytes()
   logx.WithContext(ctx).Debugf("compute account leaf hash,l1Address=%s,pk=%s,nonce=%d,collectionNonce=%d,assetRoot=%s,hash=%s",
   l1Address, pk, nonce, collectionNonce, common.Bytes2Hex(assetRoot), common.Bytes2Hex(hash[:]))
   return hash[:], nil
}
```

##### AssetTree

`AssetTree` is sub-tree of `AccountTree` and it stores all the assets `balance`, and `offerCanceledOrFinalized`. The node of asset tree is:

```go
type AssetNode struct {
	Balance  string
    OfferCanceledOrFinalized string // uint128
}
```

Leaf hash computation:

```go
func ComputeAccountAssetLeafHash(
   balance string,
   offerCanceledOrFinalized string,
   ctx context.Context,
) (hashVal []byte, err error) {
   balanceBigInt, isValid := new(big.Int).SetString(balance, 10)
   if !isValid {
   return nil, zkbnbtypes.AppErrInvalidBalanceString
   }
   e0 := txtypes.FromBigIntToFr(balanceBigInt)
   
   offerCanceledOrFinalizedBigInt, isValid := new(big.Int).SetString(offerCanceledOrFinalized, 10)
   if !isValid {
   return nil, zkbnbtypes.AppErrInvalidBalanceString
   }
   e1 := txtypes.FromBigIntToFr(offerCanceledOrFinalizedBigInt)
   ele := GMimcElements([]*fr.Element{e0, e1})
   hash := ele.Bytes()
   logx.WithContext(ctx).Debugf("compute account asset leaf hash,balance=%s,offerCanceledOrFinalized=%s,hash=%s",
   balance, offerCanceledOrFinalized, common.Bytes2Hex(hash[:]))
   return hash[:], nil
}
```

#### NftTree

`NftTree` is used for storing all the NFTs and the node info is:

```go
type NftNode struct {
    CreatorAccountIndex int64
    OwnerAccountIndex   int64
    NftContentHash      string
    RoyaltyRate         int64
    CollectionId        int64 // 32 bit
}
```

Leaf hash computation:

```go
func ComputeNftAssetLeafHash(
   creatorAccountIndex int64,
   ownerAccountIndex int64,
   nftContentHash string,
   royaltyRate int64,
   collectionId int64,
   ctx context.Context, 
) (hashVal []byte, err error) {
   e0 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(creatorAccountIndex))
   e1 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(ownerAccountIndex))
   
   var e2 *fr.Element
   var e3 *fr.Element
   contentHash := common.Hex2Bytes(nftContentHash)
   if len(contentHash) >= types.NftContentHashBytesSize {
   e2, err = txtypes.FromBytesToFr(contentHash[:types.NftContentHashBytesSize])
   e3, err = txtypes.FromBytesToFr(contentHash[types.NftContentHashBytesSize:])
   } else {
   e2, err = txtypes.FromBytesToFr(contentHash[:])
   }
   if err != nil {
   return nil, err
   }
   
   e4 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(royaltyRate))
   e5 := txtypes.FromBigIntToFr(new(big.Int).SetInt64(collectionId))
   var hash [32]byte
   if e3 != nil {
   ele := GMimcElements([]*fr.Element{e0, e1, e2, e3, e4, e5})
   hash = ele.Bytes()
   } else {
   ele := GMimcElements([]*fr.Element{e0, e1, e2, e4, e5})
   hash = ele.Bytes()
   }
   logx.WithContext(ctx).Debugf("compute nft asset leaf hash,creatorAccountIndex=%d,ownerAccountIndex=%d,nftContentHash=%s,royaltyRate=%d,collectionId=%d,hash=%s",
   creatorAccountIndex, ownerAccountIndex, nftContentHash, royaltyRate, collectionId, common.Bytes2Hex(hash[:]))
   
   return hash[:], nil
}
```

#### StateRoot

`StateRoot` is the final root that shows the final layer-2 state and will be stored on L1. It is computed by the root of `AccountTree` and `NftTree`. The computation of `StateRoot` works as follows:

`StateRoot = MiMC(AccountRoot || NftRoot)`

```go
func ComputeStateRootHash(
   accountRoot []byte,
   nftRoot []byte,
) []byte {
   e0 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(accountRoot))
   e1 := txtypes.FromBigIntToFr(new(big.Int).SetBytes(nftRoot))
   
   ele := GMimcElements([]*fr.Element{e0, e1})
   hash := ele.Bytes()
   return hash[:]
}
```

## ZkBNB Transactions

ZkBNB transactions are divided into Rollup transactions (initiated inside Rollup by a Rollup account) and Priority operations (initiated on the BSC by an BNB Smart Chain account).

Rollup transactions:

- EmptyTx
- ChangePubKey
- Transfer
- Withdraw
- CreateCollection
- MintNft
- TransferNft
- AtomicMatch
- CancelOffer
- WithdrawNft

Priority operations:


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

### ChangePubKey

#### Description

This is a layer-2 transaction and a user needs to call this method first to register a layer-2 account.

| Name              | Size(byte) | Comment                        |
|-------------------|------------|--------------------------------|
| TxType            | 1          | transaction type               |
| AccountIndex      | 4          | unique account index           |
| L1Address         | 20         | L1Address                      |
| Nonce             | 4          | Nonce                          |
| PubKeyX           | 32         | layer-2 account's public key X |
| PubKeyY           | 32         | layer-2 account's public key Y |
| GasFeeAssetId     | 2          | gas fee asset id               |
| GasFeeAssetAmount | 2          | packed fee amount              |
| GasAccountIndex   | 4          | gas account index              |
| ExpiredAt         | 4          | expired at                     |
| Sig               |            | l2 sig                         |
| L1Sig             |            | l1 sig                         |

```go
type ChangePubKeyInfo struct {
	AccountIndex      int64
	L1Address         string
	Nonce             int64
	PubKeyX           []byte
	PubKeyY           []byte
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	ExpiredAt         int64
	Sig               []byte
	L1Sig             string
}
```

#### 1.1: API Input sign Json

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 97                |

##### Pub Data Structure

| Name              | Size(byte) | Comment                        |
|-------------------|------------|--------------------------------|
| TxType            | 1          | transaction type               |
| AccountIndex      | 4          | unique account index           |
| pubKeyX           | 32         | layer-2 account's public key X |
| pubKeyY           | 32         | layer-2 account's public key Y |
| l1Address         | 20         | account's l1Address            |
| nonce             | 4          | layer-2 account's nonce        |
| gasFeeAssetId     | 2          | gas fee asset id               |
| gasFeeAssetAmount | 2          | packed fee amount              |

```go
func ConvertTxToChangePubKeyPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseChangePubKeyTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse change pub key tx failed: %s", err.Error())
   return nil, errors.New("invalid tx info")
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeChangePubKey))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   // because we can get Y from X, so we only need to store X is enough
   buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.PubKeyX))
   buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.PubKeyY))
   buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.Nonce)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   return err
   }
   buf.Write(packedFeeBytes)
   return buf.Bytes(), nil
}
```

#### User transaction

Input form data
```go
tx_type:1
tx_info:{
"AccountIndex": 2,
"L1Address": "0xB64d00616958131824B472CC20C3d47Bb5d9926C",
"Nonce": 7,
"PubKeyX": "EG/uk14D7iEZVvdzQwnCY1ad25IbDa4fKBlZrOfZdc4=",
"PubKeyY": "Kf/flLyPhKeDjoPvx8g1ceKGqow1WSKTjaFTv/EfXtc=",
"GasAccountIndex": 1,
"GasFeeAssetId": 0,
"GasFeeAssetAmount": 10000000000000,
"ExpiredAt": 1679310615292,
"Sig": "r+V7gaMQU0+/WbAlGG76Hb7ilHjP9znR/KzO+4vUj5gEIFcdzjpJWVsRwS9WN1tB1kI7s2JJKoNcuiHs87cVwg==",
"L1Sig": "0x8c9a4fa4901e7c15b56cdfceadd3493050a9b55f76faf1a1421db3a2aaf17bcc2118d84f32bcf1976176a84e129e8a232a08ec9948480aebb3c40240e841c1de1b"
}
```

Signed transaction representation.

L1 Signed transaction fields:

```go
signature.SignatureTemplateChangePubKey
common.Bytes2Hex(txInfo.PubKeyX),
common.Bytes2Hex(txInfo.PubKeyY), 
signature.GetHex10FromInt64(txInfo.Nonce), 
signature.GetHex10FromInt64(txInfo.AccountIndex)
```

L2 Signed transaction fields:
```go
ChainId
txType
txInfo.AccountIndex
txInfo.Nonce
txInfo.ExpiredAt
txInfo.GasFeeAssetId
packedFee
txInfo.L1Address
txInfo.PubKeyX
txInfo.PubKeyY
```

#### Circuit

```go
func VerifyChangePubKeyTx(
   api API, flag Variable,
   tx *ChangePubKeyTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   pubData = CollectPubDataFromChangePubKey(api, *tx)
   //CheckEmptyAccountNode(api, flag, accountsBefore[0])
   // verify params
   // account index
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
   // l1 address
   IsVariableEqual(api, flag, tx.L1Address, accountsBefore[0].L1Address)
   // nonce
   IsVariableEqual(api, flag, tx.Nonce, accountsBefore[0].Nonce)
   // asset id
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[0].AssetsInfo[0].AssetId)
   // should have enough assets
   tx.GasFeeAssetAmount = UnpackAmount(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
return pubData
}

```

### Deposit

#### Description

This is a layer-1 transaction and is used for depositing assets into the layer-2 account.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 43                |



##### Structure

| Name         | Size(byte) | Comment             |     |
|--------------|------------|---------------------|-----|
| TxType       | 1          | transaction type    |     |
| AccountIndex | 4          | account index       |     |
| AssetId      | 2          | asset index         |     |
| AssetAmount  | 16         | state amount        |     |
| L1Address    | 20         | account's l1Address |     |

```go
type DepositTxInfo struct {
	TxType uint8
	// Get from layer1 events.
    L1Address string
	AssetId         int64
	AssetAmount     *big.Int
	// Set by layer2.
	AccountIndex int64
}
```

```go
func ConvertTxToDepositPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseDepositTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse deposit tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeDeposit))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
   buf.Write(common2.Uint128ToBytes(txInfo.AssetAmount))
   return buf.Bytes(), nil
}
```

#### User transaction

##### DepositBNB

| Name        | Size(byte) | Comment             |
|-------------|------------|---------------------|
| L1Address   | 20         | account's l1Address |

##### DepositBEP20

| Name         | Size(byte)  | Comment               |
|--------------|-------------|-----------------------|
| AssetAddress | 20          | asset layer-1 address |
| Amount       | 13          | asset layer-1 amount  |
| L1Address    | 20          | account's l1Address   |

#### Circuit

```go
func VerifyDepositTx(
   api API, flag Variable,
   tx DepositTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   pubData = CollectPubDataFromDeposit(api, tx)
   // verify params
   isNewAccount := api.IsZero(api.Cmp(accountsBefore[0].L1Address, ZeroInt))
   address := api.Select(isNewAccount, tx.L1Address, accountsBefore[0].L1Address)
   IsVariableEqual(api, flag, tx.L1Address, address)
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
|--------|-------------------|
|        | 71                |

##### Structure
```go
type DepositNftTxInfo struct {
	TxType uint8

	// Get from layer1 events.
    L1Address           string
	CreatorAccountIndex int64
	CreatorTreasuryRate int64
	NftContentHash      []byte
	CollectionId        int64

	// New nft set by layer2, otherwise get from layer1.
	NftIndex int64

	// Set by layer2.
	AccountIndex int64
}
```

| Name                | Size(byte) | Comment               |
|---------------------|------------|-----------------------|
| TxType              | 1          | transaction type      |
| AccountIndex        | 4          | account index         |
| NftIndex            | 5          | unique index of a nft |
| CreatorAccountIndex | 4          | creator account index |
| RoyaltyRate         | 2          | creator treasury rate |
| CollectionId        | 2          | collection id         |
| NftContentHash      | 32         | nft content hash      |
| NftContentType      | 1          | nft content type      |
| L1Address           | 20         | account's l1Address   |

```go
func ConvertTxToDepositNftPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseDepositNftTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse deposit nft tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeDepositNft))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
   buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
   buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
   buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
   buf.WriteByte(uint8(txInfo.NftContentType))
   return buf.Bytes(), nil
}
```

#### User transaction

| Name         | Size(byte) | Comment                      |
|--------------|------------|------------------------------|
| L1Address    | 20         | account's l1Address          |
| AssetAddress | 20         | nft contract layer-1 address |
| NftTokenId   | 32         | nft layer-1 token id         |

#### Circuit

```go
func VerifyDepositNftTx(
   api API,
   flag Variable,
   tx DepositNftTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
   nftBefore NftConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   pubData = CollectPubDataFromDepositNft(api, tx)
   
   // verify params
   // check empty nft
   CheckEmptyNftNode(api, flag, nftBefore)
   // account index
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
   // account address
   isNewAccount := api.IsZero(api.Cmp(accountsBefore[0].L1Address, ZeroInt))
   address := api.Select(isNewAccount, tx.L1Address, accountsBefore[0].L1Address)
   IsVariableEqual(api, flag, tx.L1Address, address)
   //NftContentType
   IsVariableEqual(api, flag, tx.NftContentType, nftBefore.NftContentType)
   return pubData
}
```

### Transfer

#### Description

This is a layer-2 transaction and is used for transferring assets in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 72                |

##### Structure

| Name              | Size(byte) | Comment                  |
|-------------------|------------|--------------------------|
| TxType            | 1          | transaction type         |
| FromAccountIndex  | 4          | from account index       |
| ToAccountIndex    | 4          | receiver account index   |
| ToL1Address       | 20         | receiver account address |
| AssetId           | 2          | asset index              |
| AssetAmount       | 5          | packed asset amount      |
| GasFeeAssetId     | 2          | gas fee asset id         |
| GasFeeAssetAmount | 2          | packed fee amount        |
| CallDataHash      | 32         | call data hash           |

```go
func ConvertTxToTransferPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseTransferTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse transfer tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeTransfer))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
   buf.Write(common2.AddressStrToBytes(txInfo.ToL1Address))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
   packedAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.AssetAmount)
   if err != nil {
   return err
   }
   buf.Write(packedAmountBytes)
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   return err
   }
   buf.Write(packedFeeBytes)
   buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
   return pubData, nil
}
```

#### User transaction

```go
type TransferTxInfo struct {
FromAccountIndex  int64
ToL1Address       string
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
L1Sig             string
}
```

#### Circuit

```go
func VerifyTransferTx(
   api API, flag Variable,
   tx *TransferTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   toAccount := 1
   
   // collect pubdata
   pubData = CollectPubDataFromTransfer(api, *tx)
   // verify params
   // account index
   IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[fromAccount].AccountIndex)
   // account to l1Address
   isNewAccount := api.IsZero(api.Cmp(accountsBefore[toAccount].L1Address, ZeroInt))
   address := api.Select(isNewAccount, tx.ToL1Address, accountsBefore[toAccount].L1Address)
   IsVariableEqual(api, flag, tx.ToL1Address, address)
   // asset id
   IsVariableEqual(api, flag, tx.AssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.AssetId, accountsBefore[toAccount].AssetsInfo[0].AssetId)
   // gas asset id
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[1].AssetId)
   // should have enough balance
   tx.AssetAmount = UnpackAmount(api, tx.AssetAmount)
   tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.AssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[1].Balance)
   return pubData
}
```

### Withdraw

#### Description

This is a layer-2 transaction and is used for withdrawing assets from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 47                |

##### Structure

| Name              | Size(byte) | Comment                  |
|-------------------|------------|--------------------------|
| TxType            | 1          | transaction type         |
| AccountIndex      | 4          | from account index       |
| ToAddress         | 20         | layer-1 receiver address |
| AssetId           | 2          | asset index              |
| AssetAmount       | 16         | state amount             |
| GasFeeAssetId     | 2          | gas fee asset id         |
| GasFeeAssetAmount | 2          | packed fee amount        |

```go
func ConvertTxToWithdrawPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseWithdrawTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse withdraw tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeWithdraw))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
   buf.Write(common2.AddressStrToBytes(txInfo.ToAddress))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
   buf.Write(common2.Uint128ToBytes(txInfo.AssetAmount))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
   return err
   }
   buf.Write(packedFeeBytes)
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
    L1Sig             string
}
```

#### Circuit

```go
func VerifyWithdrawTx(
   api API, flag Variable,
   tx *WithdrawTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   pubData = CollectPubDataFromWithdraw(api, *tx)
   // verify params
   // account index
   IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[fromAccount].AccountIndex)
   // asset id
   IsVariableEqual(api, flag, tx.AssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[1].AssetId)
   // should have enough assets
   tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.AssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[1].Balance)
   return pubData
}
```

### CreateCollection

#### Description

This is a layer-2 transaction and is used for creating a new collection

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 11                |

##### Structure

| Name              | Size(byte) | Comment           |
|-------------------|------------|-------------------|
| TxType            | 1          | transaction type  |
| AccountIndex      | 4          | account index     |
| CollectionId      | 2          | collection index  |
| GasFeeAssetId     | 2          | asset id          |
| GasFeeAssetAmount | 2          | packed fee amount |

```go
func ConvertTxToCreateCollectionPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseCreateCollectionTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse transfer tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeCreateCollection))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
   return err
   }
   buf.Write(packedFeeBytes)
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
    L1Sig             string
}
```

#### Circuit

```go
func VerifyCreateCollectionTx(
   api API, flag Variable,
   tx *CreateCollectionTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   pubData = CollectPubDataFromCreateCollection(api, *tx)
   // verify params
   IsVariableLessOrEqual(api, flag, tx.CollectionId, 65535)
   // account index
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[fromAccount].AccountIndex)
   // asset id
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   // collection id
   IsVariableEqual(api, flag, tx.CollectionId, accountsBefore[fromAccount].CollectionNonce)
   // should have enough assets
   tx.GasFeeAssetAmount = UnpackAmount(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   return pubData
}
```

### MintNft

#### Description

This is a layer-2 transaction and is used for minting NFTs in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 75                |

##### Structure

| Name                | Size(byte) | Comment                  |
|---------------------|------------|--------------------------|
| TxType              | 1          | transaction type         |
| CreatorAccountIndex | 4          | creator account index    |
| ToAccountIndex      | 4          | receiver account index   |
| ToL1Address         | 20         | receiver account address |
| NftIndex            | 5          | unique nft index         |
| GasFeeAssetId       | 2          | gas fee asset id         |
| GasFeeAssetAmount   | 2          | packed fee amount        |
| RoyaltyRate         | 2          | creator treasury rate    |
| CollectionId        | 2          | collection index         |
| NftContentHash      | 32         | nft content hash         |
| NftContentType      | 1          | nft content type         |

```go
func ConvertTxToMintNftPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseMintNftTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse transfer tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeMintNft))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
   buf.Write(common2.AddressStrToBytes(txInfo.ToL1Address))
   buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
   return err
   }
   buf.Write(packedFeeBytes)
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.NftCollectionId)))
   buf.Write(common2.PrefixPaddingBufToChunkSize(common.FromHex(txInfo.NftContentHash)))
   buf.WriteByte(uint8(txInfo.NftContentType))
   return buf.Bytes(), nil
}
```

#### User transaction

```go
type MintNftTxInfo struct {
	CreatorAccountIndex int64
	ToAccountIndex      int64
    ToL1Address         string
	NftIndex            int64
	NftContentHash      string
	NftCollectionId     int64
	CreatorTreasuryRate int64
	GasAccountIndex     int64
	GasFeeAssetId       int64
	GasFeeAssetAmount   *big.Int
	ExpiredAt           int64
	Nonce               int64
    MetaData            string
    MutableAttributes   string
	Sig                 []byte
    L1Sig               string
}
```

#### Circuit

```go
func VerifyMintNftTx(
   api API, flag Variable,
   tx *MintNftTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints, nftBefore NftConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   toAccount := 1
   
   pubData = CollectPubDataFromMintNft(api, *tx)
   // verify params
   // check empty nft
   CheckEmptyNftNode(api, flag, nftBefore)
   // account index
   IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[fromAccount].AccountIndex)
   IsVariableEqual(api, flag, tx.ToAccountIndex, accountsBefore[toAccount].AccountIndex)
   // account address
   // account to l1Address
   isNewAccount := api.IsZero(api.Cmp(accountsBefore[toAccount].L1Address, ZeroInt))
   address := api.Select(isNewAccount, tx.ToL1Address, accountsBefore[toAccount].L1Address)
   IsVariableEqual(api, flag, tx.ToL1Address, address)
   // content hash
   isZero := api.Or(api.IsZero(tx.NftContentHash[0]), api.IsZero(tx.NftContentHash[1]))
   IsVariableEqual(api, flag, isZero, 0)
   // gas asset id
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   // should have enough balance
   tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   // collection id should be less than creator's collection nonce
   IsVariableLess(api, flag, tx.CollectionId, accountsBefore[fromAccount].CollectionNonce)
   //NftContentType
   IsVariableLessOrEqual(api, flag, 0, tx.NftContentType)
return pubData
}
```

### TransferNft

#### Description

This is a layer-2 transaction and is used for transferring NFTs to others in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 70                |

##### Structure

| Name              | Size(byte) | Comment                  |
|-------------------|------------|--------------------------|
| TxType            | 1          | transaction type         |
| FromAccountIndex  | 4          | from account index       |
| ToAccountIndex    | 4          | receiver account index   |
| ToL1Address       | 20         | receiver account address |
| NftIndex          | 5          | unique nft index         |
| GasFeeAssetId     | 2          | gas fee asset id         |
| GasFeeAssetAmount | 2          | packed fee amount        |
| CallDataHash      | 32         | call data hash           |

```go
func ConvertTxToTransferNftPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseTransferNftTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse transfer tx failed: %s", err.Error())
   return nil, errors.New("invalid tx info")
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeTransferNft))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
   buf.Write(common2.AddressStrToBytes(txInfo.ToL1Address))
   buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   return err
   }
   buf.Write(packedFeeBytes)
   buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
   return buf.Bytes(), nil
}
```

#### User transaction

```go
type TransferNftTxInfo struct {
	FromAccountIndex  int64
    ToL1Address       string
	NftIndex          int64
	GasAccountIndex   int64
	GasFeeAssetId     int64
	GasFeeAssetAmount *big.Int
	CallData          string
	CallDataHash      []byte
	ExpiredAt         int64
	Nonce             int64
	Sig               []byte
    L1Sig             string
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
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   toAccount := 1
   pubData = CollectPubDataFromTransferNft(api, *tx)
   // verify params
   // account index
   IsVariableEqual(api, flag, tx.FromAccountIndex, accountsBefore[fromAccount].AccountIndex)
   // account address
   isNewAccount := api.IsZero(api.Cmp(accountsBefore[toAccount].L1Address, ZeroInt))
   address := api.Select(isNewAccount, tx.ToL1Address, accountsBefore[toAccount].L1Address)
   IsVariableEqual(api, flag, tx.ToL1Address, address)
   // asset id
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   // nft info
   IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
   IsVariableEqual(api, flag, tx.FromAccountIndex, nftBefore.OwnerAccountIndex)
   // should have enough balance
   tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   return pubData
}
```

### AtomicMatch

#### Description

This is a layer-2 transaction that will be used for buying or selling Nft in the layer-2 network.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 63                |

##### Structure

`Offer`:

| Name                | Size(byte) | Comment                                                                                       |
|---------------------|------------|-----------------------------------------------------------------------------------------------|
| Type                | 1          | transaction type, 0 indicates this is a  `BuyNftOffer` , 1 indicate this is a  `SellNftOffer` |
| OfferId             | 3          | used to identify the oﬀer                                                                     |
| AccountIndex        | 4          | who want to buy/sell nft                                                                      |
| AssetId             | 2          | the asset id which buyer/seller want to use pay for nft                                       |
| AssetAmount         | 5          | the asset amount                                                                              |
| ListedAt            |            | timestamp when the order is signed                                                            |
| ExpiredAt           |            | timestamp after which the order is invalid                                                    |
| ChannelAccountIndex | 4          | channel account index                                                                         |
| ChannelRate         | 5          | channel rate                                                                                  |
| ProtocolRate        |            | protocol rate                                                                                 |
| ProtocolAmount      | 5          | protocol amount                                                                               |
| Sig                 |            | signature generated by buyer/seller_account_index's private key                               |
| L1Sig               |            | l1 signature generated by buyer/seller_account_index's private key                            |

`AtomicMatch`(**below is the only info that will be uploaded on-chain**):

| Name                    | Size(byte) | Comment                    |
|-------------------------|------------|----------------------------|
| TxType                  | 1          | transaction type           |
| SubmitterAccountIndex   | 4          | submitter account index    |
| BuyerAccountIndex       | 4          | buyer account index        |
| BuyerOfferId            | 3          | used to identify the offer |
| SellerAccountIndex      | 4          | seller account index       |
| SellerOfferId           | 3          | used to identify the offer |
| NftIndex                | 5          | nft id                     |
| AssetId                 | 2          | asset id                   |
| AssetAmount             | 5          | packed asset amount        |
| RoyaltyAmount           | 5          | packed creator amount      |
| GasFeeAssetId           | 2          | gas fee asset id           |
| GasFeeAssetAmount       | 2          | packed fee amount          |
| BuyProtocolAmount       | 5          | packed buy protocol amount |
| BuyChannelAccountIndex  | 4          | buy  ChannelAccountIndex   |
| BuyChannelAmount        | 5          | buy  ChannelAmount         |
| SellChannelAccountIndex | 4          | sell  ChannelAccountIndex  |
| SellChannelAmount       | 5          | sell  ChannelAmount        |


```go
func ConvertTxToAtomicMatchPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseAtomicMatchTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse atomic match tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeAtomicMatch))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.BuyOffer.AccountIndex)))
   buf.Write(common2.Uint24ToBytes(txInfo.BuyOffer.OfferId))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.SellOffer.AccountIndex)))
   buf.Write(common2.Uint24ToBytes(txInfo.SellOffer.OfferId))
   buf.Write(common2.Uint40ToBytes(txInfo.BuyOffer.NftIndex))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.SellOffer.AssetId)))
   packedAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.BuyOffer.AssetAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
   return err
   }
   buf.Write(packedAmountBytes)
   
   royaltyAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.RoyaltyAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
   return err
   }
   buf.Write(royaltyAmountBytes)
   
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
   return err
   }
   buf.Write(packedFeeBytes)
   
   protocolAmountBytes, err := common2.AmountToPackedAmountBytes(txInfo.BuyOffer.ProtocolAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
   return err
   }
   buf.Write(protocolAmountBytes)
   
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.BuyOffer.ChannelAccountIndex)))
   buyChanelAmount, err := common2.AmountToPackedAmountBytes(txInfo.BuyChannelAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
   return err
   }
   buf.Write(buyChanelAmount)
   
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.SellOffer.ChannelAccountIndex)))
   sellChanelAmount, err := common2.AmountToPackedAmountBytes(txInfo.SellChannelAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
   return err
   }
   buf.Write(sellChanelAmount)
   return buf.Bytes(), nil
}
```

#### User transaction

```go
type OfferTxInfo struct {
   Type                int64
   OfferId             int64
   AccountIndex        int64
   NftIndex            int64
   NftName             string
   AssetId             int64
   AssetName           string
   AssetAmount         *big.Int
   ListedAt            int64
   ExpiredAt           int64
   RoyaltyRate         int64
   ChannelAccountIndex int64
   ChannelRate         int64
   ProtocolRate        int64
   ProtocolAmount      *big.Int
   Sig                 []byte
   L1Sig               string
}

type AtomicMatchTxInfo struct {
   AccountIndex      int64
   BuyOffer          *OfferTxInfo
   SellOffer         *OfferTxInfo
   GasAccountIndex   int64
   GasFeeAssetId     int64
   GasFeeAssetAmount *big.Int
   RoyaltyAmount     *big.Int
   BuyChannelAmount  *big.Int
   SellChannelAmount *big.Int
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
) (pubData [PubDataBitsSizePerTx]Variable, err error) {
   fromAccount := 0
   buyAccount := 1
   sellAccount := 2
   creatorAccount := 3
   buyChanelAccount := 4
   sellChanelAccount := 5
   protocolAccount := 6
   
   pubData = CollectPubDataFromAtomicMatch(api, *tx)
   // verify params
   IsVariableEqual(api, flag, tx.BuyOffer.Type, 0)
   IsVariableEqual(api, flag, tx.SellOffer.Type, 1)
   IsVariableEqual(api, flag, tx.BuyOffer.AssetId, tx.SellOffer.AssetId)
   IsVariableEqual(api, flag, tx.BuyOffer.AssetAmount, tx.SellOffer.AssetAmount)
   IsVariableEqual(api, flag, tx.BuyOffer.NftIndex, tx.SellOffer.NftIndex)
   IsVariableEqual(api, flag, tx.BuyOffer.AssetId, accountsBefore[buyAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.BuyOffer.AssetId, accountsBefore[creatorAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.BuyOffer.AssetId, accountsBefore[buyChanelAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.BuyOffer.AssetId, accountsBefore[sellChanelAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.BuyOffer.AssetId, accountsBefore[protocolAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.SellOffer.AssetId, accountsBefore[sellAccount].AssetsInfo[0].AssetId)
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   IsVariableLessOrEqual(api, flag, blockCreatedAt, tx.BuyOffer.ExpiredAt)
   IsVariableLessOrEqual(api, flag, blockCreatedAt, tx.SellOffer.ExpiredAt)
   IsVariableEqual(api, flag, nftBefore.NftIndex, tx.SellOffer.NftIndex)
   IsVariableEqual(api, flag, nftBefore.RoyaltyRate, tx.BuyOffer.RoyaltyRate)
   
   // verify signature
   buyOfferHash := ComputeHashFromBuyOfferTx(api, tx.BuyOffer)
   notBuyer := api.IsZero(api.IsZero(api.Sub(tx.AccountIndex, tx.BuyOffer.AccountIndex)))
   notBuyer = api.And(flag, notBuyer)
   err = VerifyEddsaSig(notBuyer, api, hFunc, buyOfferHash, accountsBefore[1].AccountPk, tx.BuyOffer.Sig)
   if err != nil {
   return pubData, err
   }
   sellOfferHash := ComputeHashFromSellOfferTx(api, tx.SellOffer)
   notSeller := api.IsZero(api.IsZero(api.Sub(tx.AccountIndex, tx.SellOffer.AccountIndex)))
   notSeller = api.And(flag, notSeller)
   err = VerifyEddsaSig(notSeller, api, hFunc, sellOfferHash, accountsBefore[2].AccountPk, tx.SellOffer.Sig)
   if err != nil {
   return pubData, err
   }
   // verify account index
   // submitter
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[fromAccount].AccountIndex)
   // buyer
   IsVariableEqual(api, flag, tx.BuyOffer.AccountIndex, accountsBefore[buyAccount].AccountIndex)
   // seller
   IsVariableEqual(api, flag, tx.SellOffer.AccountIndex, accountsBefore[sellAccount].AccountIndex)
   // creator
   IsVariableEqual(api, flag, nftBefore.CreatorAccountIndex, accountsBefore[creatorAccount].AccountIndex)
   // buyChanelAccount
   IsVariableEqual(api, flag, tx.BuyOffer.ChannelAccountIndex, accountsBefore[buyChanelAccount].AccountIndex)
   // sellChanelAccount
   IsVariableEqual(api, flag, tx.SellOffer.ChannelAccountIndex, accountsBefore[sellChanelAccount].AccountIndex)
   // sellChanelAccount
   IsVariableEqual(api, flag, tx.ProtocolAccountIndex, accountsBefore[protocolAccount].AccountIndex)
   
   // verify buy offer id
   buyOfferIdBits := api.ToBinary(tx.BuyOffer.OfferId, 24)
   buyAssetId := api.FromBinary(buyOfferIdBits[7:]...)
   buyOfferIndex := api.Sub(tx.BuyOffer.OfferId, api.Mul(buyAssetId, OfferSizePerAsset))
   buyOfferIndexBits := api.ToBinary(accountsBefore[buyAccount].AssetsInfo[1].OfferCanceledOrFinalized, OfferSizePerAsset)
   for i := 0; i < OfferSizePerAsset; i++ {
   isZero := api.IsZero(api.Sub(buyOfferIndex, i))
   IsVariableEqual(api, isZero, buyOfferIndexBits[i], 0)
   }
   // verify sell offer id
   sellOfferIdBits := api.ToBinary(tx.SellOffer.OfferId, 24)
   sellAssetId := api.FromBinary(sellOfferIdBits[7:]...)
   sellOfferIndex := api.Sub(tx.SellOffer.OfferId, api.Mul(sellAssetId, OfferSizePerAsset))
   sellOfferIndexBits := api.ToBinary(accountsBefore[sellAccount].AssetsInfo[1].OfferCanceledOrFinalized, OfferSizePerAsset)
   for i := 0; i < OfferSizePerAsset; i++ {
   isZero := api.IsZero(api.Sub(sellOfferIndex, i))
   IsVariableEqual(api, isZero, sellOfferIndexBits[i], 0)
   }
   // buyer should have enough balance
   tx.BuyOffer.AssetAmount = UnpackAmount(api, tx.BuyOffer.AssetAmount)
   tx.BuyOffer.ProtocolAmount = UnpackAmount(api, tx.BuyOffer.ProtocolAmount)
   tx.BuyChannelAmount = UnpackAmount(api, tx.BuyChannelAmount)
   tx.RoyaltyAmount = UnpackAmount(api, tx.RoyaltyAmount)
   totalAmount := api.Add(tx.BuyOffer.AssetAmount, tx.BuyOffer.ProtocolAmount, tx.BuyChannelAmount, tx.RoyaltyAmount)
   IsVariableLessOrEqual(api, flag, totalAmount, accountsBefore[buyAccount].AssetsInfo[0].Balance)
   // submitter should have enough balance
   tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   
   // verify protocol amount
   protocolAmount := api.Mul(tx.BuyOffer.AssetAmount, tx.BuyOffer.ProtocolRate)
   protocolAmount = api.Div(protocolAmount, RateBase)
   IsVariableEqual(api, flag, tx.BuyOffer.ProtocolAmount, protocolAmount)
   
   // verify royalty amount
   royaltyAmount := api.Mul(tx.BuyOffer.AssetAmount, tx.BuyOffer.RoyaltyRate)
   royaltyAmount = api.Div(royaltyAmount, RateBase)
   IsVariableEqual(api, flag, tx.RoyaltyAmount, royaltyAmount)
   return pubData, nil
}
```

### CancelOffer

#### Description

This is a layer-2 transaction and is used for canceling nft offer.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 12                |

##### Structure

| Name              | Size(byte) | Comment           |
|-------------------|------------|-------------------|
| TxType            | 1          | transaction type  |
| AccountIndex      | 4          | account index     |
| OfferId           | 3          | nft offer id      |
| GasFeeAssetId     | 2          | gas fee asset id  |
| GasFeeAssetAmount | 2          | packed fee amount |

```go
func ConvertTxToCancelOfferPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseCancelOfferTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse transfer tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeCancelOffer))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.Uint24ToBytes(txInfo.OfferId))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
   return err
   }
   buf.Write(packedFeeBytes)
   return buf.Bytes(), nil
}
```

#### User transaction

```go
type CancelOfferTxInfo struct {
   AccountIndex      int64
   OfferId           int64
   NftName           string
   GasAccountIndex   int64
   GasFeeAssetId     int64
   GasFeeAssetAmount *big.Int
   ExpiredAt         int64
   Nonce             int64
   Sig               []byte
   L1Sig             string
}
```

#### Circuit

```go
func VerifyCancelOfferTx(
   api API, flag Variable,
   tx *CancelOfferTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   pubData = CollectPubDataFromCancelOffer(api, *tx)
   // verify params
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[fromAccount].AccountIndex)
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   offerIdBits := api.ToBinary(tx.OfferId, 24)
   assetId := api.FromBinary(offerIdBits[7:]...)
   IsVariableEqual(api, flag, assetId, accountsBefore[fromAccount].AssetsInfo[1].AssetId)
   // should have enough balance
   tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   return pubData
}
```

### WithdrawNft

#### Description

This is a layer-2 transaction and is used for withdrawing nft from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
| 0      | 95                |

##### Structure

| Name                | Size(byte) | Comment               |
|---------------------|------------|-----------------------|
| TxType              | 1          | transaction type      |
| AccountIndex        | 4          | account index         |
| CreatorAccountIndex | 4          | creator account index |
| RoyaltyRate         | 2          | creator treasury rate |
| NftIndex            | 5          | unique nft index      |
| CollectionId        | 2          | collection id         |
| GasFeeAssetId       | 2          | gas fee asset id      |
| GasFeeAssetAmount   | 2          | packed fee amount     |
| ToAddress           | 20         | receiver address      |
| CreatorL1Address    | 20         | creatot l1Address     |
| NftContentHash      | 32         | nft content hash      |
| NftContentType      | 1          | nft content type      |

```go
func ConvertTxToWithdrawNftPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseWithdrawNftTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse transfer tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeWithdrawNft))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
   buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
   packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
   if err != nil {
   logx.Errorf("unable to convert amount to packed fee amount: %s", err.Error())
   return err
   }
   buf.Write(packedFeeBytes)
   buf.Write(common2.AddressStrToBytes(txInfo.ToAddress))
   buf.Write(common2.AddressStrToBytes(txInfo.CreatorL1Address))
   buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
   buf.WriteByte(uint8(txInfo.NftContentType))
   return buf.Bytes(), nil
}
```

#### User transaction

```go
type WithdrawNftTxInfo struct {
   AccountIndex        int64
   CreatorAccountIndex int64
   CreatorL1Address    string
   RoyaltyRate         int64
   NftIndex            int64
   NftName             string
   NftContentHash      []byte
   NftContentType      int64
   CollectionId        int64
   ToAddress           string
   GasAccountIndex     int64
   GasFeeAssetId       int64
   GasFeeAssetAmount   *big.Int
   ExpiredAt           int64
   Nonce               int64
   Sig                 []byte
   L1Sig               string
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
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   creatorAccount := 1
   pubData = CollectPubDataFromWithdrawNft(api, *tx)
   // verify params
   // account index
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[fromAccount].AccountIndex)
   IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[creatorAccount].AccountIndex)
   // account name hash
   IsVariableEqual(api, flag, tx.CreatorL1Address, accountsBefore[creatorAccount].L1Address)
   // collection id
   IsVariableEqual(api, flag, tx.CollectionId, nftBefore.CollectionId)
   //NftContentType
   IsVariableEqual(api, flag, tx.NftContentType, nftBefore.NftContentType)
   // asset id
   IsVariableEqual(api, flag, tx.GasFeeAssetId, accountsBefore[fromAccount].AssetsInfo[0].AssetId)
   // nft info
   IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
   IsVariableEqual(api, flag, tx.CreatorAccountIndex, nftBefore.CreatorAccountIndex)
   IsVariableEqual(api, flag, tx.RoyaltyRate, nftBefore.RoyaltyRate)
   IsVariableEqual(api, flag, tx.AccountIndex, nftBefore.OwnerAccountIndex)
   IsVariableEqual(api, flag, tx.NftContentHash[0], nftBefore.NftContentHash[0])
   IsVariableEqual(api, flag, tx.NftContentHash[1], nftBefore.NftContentHash[1])
   // have enough assets
   tx.GasFeeAssetAmount = UnpackFee(api, tx.GasFeeAssetAmount)
   IsVariableLessOrEqual(api, flag, tx.GasFeeAssetAmount, accountsBefore[fromAccount].AssetsInfo[0].Balance)
   return pubData
}
```

### FullExit

#### Description

This is a layer-1 transaction and is used for full exit assets from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 43                |

##### Structure

| Name         | Size(byte) | Comment             |
|--------------|------------|---------------------|
| TxType       | 1          | transaction type    |
| AccountIndex | 4          | from account index  |
| AssetId      | 2          | asset index         |
| AssetAmount  | 16         | state amount        |
| L1Address    | 20         | account's l1Address |
```go
type FullExitTxInfo struct {
   TxType uint8
   
   // Get from layer1 events.
   L1Address    string
   AssetId      int64
   AccountIndex int64
   
   // Set by layer2.
   AssetAmount *big.Int
}
```
```go
func ConvertTxToFullExitPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseFullExitTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse full exit tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeFullExit))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.AssetId)))
   buf.Write(common2.Uint128ToBytes(txInfo.AssetAmount))
   buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
   return buf.Bytes(), nil
}
```

#### User transaction

| Name         | Size(byte) | Comment               |
|--------------|------------|-----------------------|
| L1Address    | 20         | account's l1Address   |
| AssetAddress | 20         | asset layer-1 address |

#### Circuit

```go
func VerifyFullExitTx(
   api API, flag Variable,
   tx FullExitTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   txInfoL1Address := api.Select(flag, tx.L1Address, ZeroInt)
   beforeL1Address := api.Select(flag, accountsBefore[0].L1Address, ZeroInt)
   isOwner := api.And(api.IsZero(api.Cmp(txInfoL1Address, beforeL1Address)), flag)
   tx.AssetAmount = api.Select(isOwner, tx.AssetAmount, ZeroInt)
   pubData = CollectPubDataFromFullExit(api, tx)
   // verify params
   IsVariableEqual(api, isOwner, tx.L1Address, accountsBefore[0].L1Address)
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[0].AccountIndex)
   IsVariableEqual(api, flag, tx.AssetId, accountsBefore[0].AssetsInfo[0].AssetId)
   
   IsVariableEqual(api, isOwner, tx.AssetAmount, accountsBefore[0].AssetsInfo[0].Balance)
   return pubData
}
```

### FullExitNft

#### Description

This is a layer-1 transaction and is used for full exit NFTs from the layer-2 to the layer-1.

#### On-Chain operation

##### Size

| Chunks | Significant bytes |
|--------|-------------------|
|        | 91                |

##### Structure

| Name                | Size(byte) | Comment                   |
|---------------------|------------|---------------------------|
| TxType              | 1          | transaction type          |
| AccountIndex        | 4          | from account index        |
| CreatorAccountIndex | 4          | creator account index     |
| RoyaltyRate         | 2          | creator treasury rate     |
| NftIndex            | 5          | unique nft index          |
| CollectionId        | 2          | collection id             |
| L1Address           | 20         | account's L1Address       |
| CreatorL1Address    | 20         | creator account l1Address |
| NftContentHash      | 32         | nft content hash          |
| NftContentType      | 1          | nft content type          |

```go
type FullExitNftTxInfo struct {
   TxType uint8
   
   // Get from layer1 events.
   NftIndex     int64
   L1Address    string
   AccountIndex int64
   // Set by layer2.
   CreatorAccountIndex int64
   RoyaltyRate         int64
   CreatorL1Address    string
   NftContentHash      []byte
   NftContentType      int64
   CollectionId        int64
}

```
```go
func ConvertTxToFullExitNftPubData(tx *tx.Tx) (pubData []byte, err error) {
   txInfo, err := types.ParseFullExitNftTxInfo(tx.TxInfo)
   if err != nil {
   logx.Errorf("parse full exit nft tx failed: %s", err.Error())
   return nil, types.AppErrInvalidTxInfo
   }
   var buf bytes.Buffer
   buf.WriteByte(uint8(types.TxTypeFullExitNft))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
   buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
   buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
   buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
   buf.Write(common2.AddressStrToBytes(txInfo.L1Address))
   buf.Write(common2.AddressStrToBytes(txInfo.CreatorL1Address))
   buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
   buf.WriteByte(uint8(txInfo.NftContentType))
   return buf.Bytes(), nil
}
```

#### User transaction

| Name      | Size(byte) | Comment             |
|-----------|------------|---------------------|
| L1Address | 20         | account's L1Address |
| NftIndex  | 5          | unique nft index    |

#### Circuit

```go
func VerifyFullExitNftTx(
   api API, flag Variable,
   tx FullExitNftTxConstraints,
   accountsBefore [NbAccountsPerTx]AccountConstraints,
   nftBefore NftConstraints,
) (pubData [PubDataBitsSizePerTx]Variable) {
   fromAccount := 0
   creatorAccount := 1
   
   txInfoL1Address := api.Select(flag, tx.L1Address, ZeroInt)
   beforeL1Address := api.Select(flag, accountsBefore[fromAccount].L1Address, ZeroInt)
   isFullExitSuccess := api.IsZero(api.Cmp(txInfoL1Address, beforeL1Address))
   isOwner := api.And(isFullExitSuccess, api.And(api.IsZero(api.Sub(tx.AccountIndex, nftBefore.OwnerAccountIndex)), flag))
   
   tx.CreatorAccountIndex = api.Select(isOwner, tx.CreatorAccountIndex, ZeroInt)
   tx.NftContentHash[0] = api.Select(isOwner, tx.NftContentHash[0], ZeroInt)
   tx.NftContentHash[1] = api.Select(isOwner, tx.NftContentHash[1], ZeroInt)
   tx.RoyaltyRate = api.Select(isOwner, tx.RoyaltyRate, ZeroInt)
   tx.CollectionId = api.Select(isOwner, tx.CollectionId, ZeroInt)
   tx.NftContentType = api.Select(isOwner, tx.NftContentType, ZeroInt)
   
   pubData = CollectPubDataFromFullExitNft(api, tx)
   // verify params
   IsVariableEqual(api, isOwner, tx.L1Address, accountsBefore[fromAccount].L1Address)
   IsVariableEqual(api, flag, tx.AccountIndex, accountsBefore[fromAccount].AccountIndex)
   IsVariableEqual(api, flag, tx.NftIndex, nftBefore.NftIndex)
   IsVariableEqual(api, flag, tx.CreatorAccountIndex, accountsBefore[creatorAccount].AccountIndex)
   IsVariableEqual(api, flag, tx.CreatorL1Address, accountsBefore[creatorAccount].L1Address)
   IsVariableEqual(api, isOwner, tx.CreatorAccountIndex, nftBefore.CreatorAccountIndex)
   IsVariableEqual(api, isOwner, tx.RoyaltyRate, nftBefore.RoyaltyRate)
   IsVariableEqual(api, isOwner, tx.NftContentHash[0], nftBefore.NftContentHash[0])
   IsVariableEqual(api, isOwner, tx.NftContentHash[1], nftBefore.NftContentHash[1])
   //NftContentType
   IsVariableEqual(api, flag, tx.NftContentType, nftBefore.NftContentType)
   return pubData
}
```

## Smart contracts API

### Rollup contract

#### Deposit BNB

Deposit BNB to Rollup - transfer BNB from user L1 address into Rollup account

```js
function depositBNB(address _to) external payable onlyActive
```

- `_to`: the receiver L1 address

#### Deposit BEP20

Deposit BEP20 assets to Rollup - transfer BEP20 assets from user L1 address into Rollup account

```js
  function depositBEP20(IERC20 _token, uint104 _amount, address _to) external onlyActive
```

- `_token`: valid BEP20 address
- `_amount`: deposit amount
- `_to`: the receiver L1 address

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
  function requestFullExit(uint32 _accountIndex, address _asset) public onlyActive
```

- `_accountIndex`: Numerical id of the account
- `_asset`: BEP20 asset address, `0` for BNB

Register full exit request to withdraw NFT tokens balance from the account. Users need to call it if they believe that their transactions are censored by the validator.

```js
  function requestFullExitNft(uint32 _accountIndex, uint32 _nftIndex) public onlyActive
```

- `_accountIndex`: Numerical id of the account
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
