# OverView

## Introduction
ZkBAS is built on ZK Rollup architecture. ZkBAS bundle (or “roll-up”) hundreds of transactions off-chain and generates
cryptographic proof. These proofs can come in the form of SNARKs (succinct non-interactive argument of knowledge) which
can prove the validity of every single transaction in the Rollup Block. It means all funds are held on the BSC,
while computation and storage are performed on BAS with less cost and fast speed.

## Problems ZkBAS solves
Today BSC is experiencing network scalability problems and the core developer has proposed to use BAS in their [Outlook 
2022](https://forum.bnbchain.org/t/bsc-development-outlook-2022/44) paper to solve this problem because these side 
chains can be designed for much higher throughput and lower gas fees. 

The [BEP100](https://github.com/bnb-chain/BEPs/pull/132/files) propose a modular framework for creating BSC-compatible 
side chains and connect them by native relayer hub. The security of native relayer hub is guaranteed by the side chain.
According to [the analysis](https://blog.chainalysis.com/reports/cross-chain-bridge-hacks-2022/) of chainalysis, bridges 
are now a top target for the hackers and attacks on bridges account for 69% of total funds stolen in 2022. ZkBAS can 
perfectly solve the problem! Thanks to zkSNARK proofs, ZkBAS share the same security as BSC does.

## ZkBAS features

ZkBAS implement the following features so far:
- **L1 security**. The ZkBAS share the same security as BSC does. Thanks to zkSNARK proofs, the security is guaranteed by
  cryptographic. Users do not have to trust any third parties or keep monitoring the Rollup blocks in order to
  prevent fraud.
- **L1<>L2 Communication**. BNB, and BEP20/BEP721/BEP1155 created on BSC or zkBAS can flow freely between BSC and zkBAS.
- **Built-in instant AMM swap**. It allows digital assets to be traded without permission and automatically by using
  liquidity pools.
- **Built-in NFT marketplace**. Developer can build marketplace for crypto collectibles and non-fungible tokens (NFTs)
  out of box on ZkBAS.
- **Fast transaction speed and faster finality**.
- **Low gas fee**. The gas token on the zkBAS can be either BEP20 or BNB.
- **"Full exit" on BSC**. The user can request this operation to withdraw funds if he thinks that his transactions
  are censored by zkBAS.

## Find More
<!--ts-->
- [ZkBAS Technology](./technology.md)
- [ZkBAS Protocol Design](./protocol.md)
- [Quick Start Tutorial](./tutorial.md)
- [Tokenomics](./tokenomics.md)
- [Storage Layout](./storage_layout.md)
- [Wallets](./wallets.md)
<!--ts-->