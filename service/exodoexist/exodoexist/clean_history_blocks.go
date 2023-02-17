package monitor

import (
	"github.com/zeromicro/go-zero/core/logx"

	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
)

func (m *Monitor) CleanHistoryBlocks() (err error) {
	latestGenericBlock, err := m.L1SyncedBlockModel.GetLatestL1SyncedBlockByType(l1syncedblock.TypeGeneric)
	if err != nil {
		return err
	}
	latestGovBlock, err := m.L1SyncedBlockModel.GetLatestL1SyncedBlockByType(l1syncedblock.TypeGovernance)
	if err != nil {
		return err
	}

	minHeight := common2.MinInt64(latestGenericBlock.L1BlockHeight, latestGovBlock.L1BlockHeight)
	keepHeight := minHeight - m.Config.ChainConfig.KeptHistoryBlocksCount
	if keepHeight <= 0 {
		return nil
	}

	logx.Infof("start to clean historical synced blocks for height less than: %d", keepHeight)
	err = m.L1SyncedBlockModel.DeleteL1SyncedBlocksForHeightLessThan(keepHeight)
	if err != nil {
		return err
	}
	logx.Infof("finish to clean historical synced blocks for height less than: %d", keepHeight)
	return nil
}
