# Tokenomics

## Supported Tokens
ZkBAS can be used to transfer BNB and BEP20 tokens. ZkBAS supports a maximum of 65535 tokens, and each supported token need to be listed on ZkBAS contract first. 

ZkBAS is not responsible for security of supported token contract. Please use any token on your own risk.

## List Token
ZkBAS use `AssetGovernance` contract to manage supported tokens. To list token on ZkBAS, call function `addAsset(address _assetAddress)` of AssetGovernance contract. 

Notice there is a listing fee for listing token, and fees are denominated using a specific token. The listing fee and fee token can be retrived by calling function `listingFee` and `listingFeeToken`. Make sure the sender that calls `addAsset` has enough fee token balance.

## Fee
In ZkBAS every transaction needs to pay fees to be included in L2 blocks, and ZkBAS must pay BNB gas to commit, verify and execute L2 blocks by sending corresponding L1 transaction. The L1 fees need to be averaged per L2 transaction which is orders of magnitude cheaper than the cost of normal BNB/BEP20 transfers.

Users can pay transaction fees in multi fee tokens supported by ZkBAS. For example, suppose ZkBAS supports BNB/USDT, when users make a transaction, users can use BNB or USDT to pay transaction fees for their own convenience.