package desertexit

import (
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/types"
)

func (m *DesertExit) CleanHistoryBlocks() (err error) {
	latestGenericBlock, err := m.L1SyncedBlockModel.GetLatestL1SyncedBlockByType(l1syncedblock.TypeGeneric)
	if err != nil && err != types.DbErrNotFound {
		return err
	}
	if err == types.DbErrNotFound {
		return nil
	}

	minHeight := latestGenericBlock.L1BlockHeight
	keepHeight := minHeight - m.Config.ChainConfig.KeptHistoryBlocksCount
	if keepHeight <= 0 {
		return nil
	}

	err = m.L1SyncedBlockModel.DeleteL1SyncedBlocksForHeightLessThan(keepHeight)
	if err != nil {
		return err
	}
	return nil
}
