package types

import (
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/nft"
)

type BlockStates struct {
	Block          *block.Block
	BlockForCommit *blockForCommit.BlockForCommit

	PendingNewAccount            []*account.Account
	PendingUpdateAccount         []*account.Account
	PendingNewAccountHistory     []*account.AccountHistory
	PendingNewLiquidity          []*liquidity.Liquidity
	PendingUpdateLiquidity       []*liquidity.Liquidity
	PendingNewLiquidityHistory   []*liquidity.LiquidityHistory
	PendingNewNft                []*nft.L2Nft
	PendingUpdateNft             []*nft.L2Nft
	PendingNewNftHistory         []*nft.L2NftHistory
	PendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory

	PendingNewOffer         []*nft.Offer
	PendingNewL2NftExchange []*nft.L2NftExchange
}
