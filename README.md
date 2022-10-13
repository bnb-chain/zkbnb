# BNB ZkRollup

![banner](./docs/assets/banner.png)


The BNB ZkRollup(ZkBNB) is an infrastructure for developers that helps them to build large scale
BSC-based apps with higher throughput and much lower or even zero transaction fees. 

ZkBNB is built on ZK Rollup architecture. ZkBNB bundle (or "roll-up") hundreds of transactions off-chain and generates
cryptographic proof. These proofs can come in the form of SNARKs (succinct non-interactive argument of knowledge) which 
can prove the validity of every single transaction in the Rollup Block. It means all funds are held on the BSC, 
while computation and storage are performed on BAS with less cost and fast speed.

ZkBNB achieves the following goals:
- **L1 security**. The ZkBNB share the same security as BSC does. Thanks to zkSNARK proofs, the security is guaranteed by
  cryptographic. Users do not have to trust any third parties or keep monitoring the Rollup blocks in order to 
  prevent fraud.
- **L1<>L2 Communication**. BNB, and BEP20/BEP721/BEP1155 created on BSC or ZkBNB can flow freely between BSC and ZkBNB.
- **Built-in NFT marketplace**. Developer can build marketplace for crypto collectibles and non-fungible tokens (NFTs) 
  out of box on ZkBNB.
- **Fast transaction speed and faster finality**.
- **Low gas fee**. The gas token on the ZkBNB can be either BEP20 or BNB.
- **"Full exit" on BSC**. The user can request this operation to withdraw funds if he thinks that his transactions 
  are censored by ZkBNB.

ZkBNB starts its development based on [Zecrey](https://github.com/bnb-chain/zecrey-legend), special thanks to
[Zecrey](https://www.zecrey.com/) team and [Nodereal](https://nodereal.io/) team for their contribution.

## Table of Content
<!--ts-->
- [Framework](#Framework)
- [Document](#Document)
- [Key Features](#Key-Features)
  + [Digital Asset Management](#Digital-Asset-Management)
  + [NFT Management and Marketplace](#NFT-Management-and-Marketplace)
  + [Native Name Service](#Native-Name-Service)
  + [Seamless L1 Wallet Management](#Seamless-L1-Wallet-Management)

- [Key Tech](#Key-Tech)
  + [Sparse Merkle Tree K-V Store](#Sparse-Merkle-Tree-KV-Store)
  + [Circuit Model](#Circuit-Model)
- [Building from Source](#Building-from-Source)
- [Dev Network Setup](#Dev-Network-Setup)
- [Testnet(coming soon)](#Testnet(coming-soon))
- [Contribution](#Contribution)
- [Related Projects](#Related-Projects)
- [Outlook](#Outlook)
<!--te-->

## Framework
![Framework](./docs/assets/Frame_work.png)

- **committer**. Committer executes transactions and produce consecutive blocks.
- **monitor**. Monitor tracks events on BSC, and translates them into **transactions** on ZkBNB.
- **witness**. Witness re-executes the transactions within the block and generates witness materials.
- **prover**. Prover generates cryptographic proof based on the witness materials.
- **sender**. The sender rollups the compressed l2 blocks to L1, and submit proof to verify it.
- **api server**. The api server is the access endpoints for most users, it provides rich data, including
  digital assets, blocks, transactions, gas fees.
- **recovery**. A tool to recover the sparse merkle tree in kv-rocks based on the state world in postgresql.


## Document
The `./docs` directory includes a lot of useful documentation. You can find detail design and tutorial [there](docs/readme.md).

## Key Features

### Digital Asset Management
The ZkBNB will serve as an alternative marketplace for issuing, using, paying and exchanging digital assets in a
decentralized manner. ZkBNB and BSC share the same token universe for BNB, BEP2 and NFT tokens. This defines:
- The same token can circulate on both networks, and flow between them bi-directionally via L1 <> L2 communication.
- The total circulation of the same token should be managed across the two networks, i.e. the total effective supply 
  of a token should be the sum of the token's total effective supply on both BSC and BC.
- The tokens can only be initially created on BSC in BEP20, then pegged to the ZkBNB. It is permissionless to peg
  token onto ZkBNB.

User can **1.deposit 2.transfer 3.withdraw** both non-fungible token and fungible token on ZkBNB.

Users enter the ZK-rollup by **depositing tokens** in the rollup's contract deployed on the BSC. The ZkBNB monitor
will track deposits and submit it as a layer2 transaction, once committer verifies the transaction, users get funds on
their account, they can start transacting by sending transactions to the committer for processing.

User can **transfer** any amount of funds to any existed accounts on ZkBNB by sending a signed transaction to the
network.

**Withdrawing** from ZkBNB to BSC is straightforward. The user initiates the withdrawal transaction, the fund will be
burned on ZkBNB. Once the transaction in the next batch been rolluped, a related amount of token will be unlocked from
rollup contract to target account. 

### NFT Management and Marketplace
We target to provide an opensource NFT marketplace for users to browse, buy, sell or create their own NFT. 
The meta-data of NFT on ZkBNB sticks to the [BSC standard](https://docs.bnbchain.org/docs/nft-metadata-standard/).
The ERC721 standard NFT can be seamlessly deposited on ZkBNB, or in reverse.

![Marketplace framework](./docs/assets/NFT_Marketplace.png)

Above diagram shows the framework of Nft Marketplace and ZkBNB. All the buy/sell offer, meta-data of NFT/Collection,
medium resources, account profiles are store in the backend of NFT marketplace, only the **contendHash**,
**ownership**, **creatorTreasuryRate** and few other fields are recorded on ZkBNB. To encourage price discovery, anyone
can place buy/sell offer in the marketplace without paying any fees since the offer is cached in the backend instead of 
being sent to the ZkBNB. Once the offer is matched, an **AtomicMatch** transaction that consist of buy and sell offer
will be sent to ZkBNB to make the trade happen. Users can also cancel an offer manually by sending a cancel offer
transaction to disable the backend cached offer.

### Native Name Service
No more copying and pasting long addresses on ZkBNB. Every account on ZkBNB gets its short name, user can use that to
store funds and receive any cryptocurrency, token, or NFT. 

### Seamless L1 Wallet Management
ZkBNB natively supports ECDSA signatures and follows [EIP712](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md)
signing structure, which means most of the Ethereum wallets can seamless support ZkBNB. There is no extra effort for BSC
users to leverage ZkBNB.

## Key Tech

### Sparse Merkle Tree KV Store 
Unlike most rollup solution to put the state tree in memory, [BAS-SMT](https://github.com/bnb-chain/zkbnb-smt/) is a versioned,
snapshottable (immutable) sparse tree for persistent data. BAS-SMT is the key factor for the massive adoption of ZkBNB.

### Circuit Model
[ZkBNB Crypto](https://github.com/bnb-chain/zkbnb-crypto) is the library that describe the proving circuit. Once
the ZK-rollup node has enough transactions, it aggregates them into a batch and compiles inputs for the proving circuit 
to compile into a succinct zk-proof.


## Building from Source

1. Install necessary tools before building, and this only need to executed by once.
```shell
make tools
```

2. Build the binary.
```shell
make build
```

## Dev Network Setup
We are preparing to set up the whole system using docker composer, it is coming soon..

## Testnet(coming soon)

## Contribution
Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet, 
and are grateful for even the smallest of fixes!

If you'd like to contribute to bsc, please fork, fix, commit and send a pull request for the maintainers to review 
and merge into the main code base. If you wish to submit more complex changes though, Start by browsing 
[new issues](https://github.com/bnb-chain/zkbnb/issues) and [BEPs](https://github.com/bnb-chain/BEPs).
If you are looking for something interesting or if you have something in your mind, there is a chance it had been discussed.

## Related Projects

- [ZkBNB Rollup Contracts](https://github.com/bnb-chain/zkbnb-contract).
- [ZkBNB Crypto](https://github.com/bnb-chain/zkbnb-crypto).
- [ZkBNB Eth RPC](https://github.com/bnb-chain/zkbnb-eth-rpc).
- [ZkBNB Go SDK](https://github.com/bnb-chain/zkbnb-go-sdk).

## Outlook
We believe that zk-Rollup Sooner or later L2 The best of the track — This is a very cheap and safe first-class 
L2 Expansion solutions. However, ZkBNB is application specific so far, this makes it difficult for developers to
build custom dApp on that, we will introduce generic programability in the future... 


