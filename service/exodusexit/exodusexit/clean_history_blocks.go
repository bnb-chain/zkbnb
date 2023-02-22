package exodusexit

import (
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
)

func (m *Monitor) CleanHistoryBlocks() (err error) {
	latestGenericBlock, err := m.L1SyncedBlockModel.GetLatestL1SyncedBlockByType(l1syncedblock.TypeGeneric)
	if err != nil {
		return err
	}
	minHeight := latestGenericBlock.L1BlockHeight
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
