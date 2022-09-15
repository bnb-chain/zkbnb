/*
 * Copyright © 2021 ZkBNB Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package monitor

import (
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	MaxRecordsToKeep = 10000
)

func (m *Monitor) Prune() {
	m.pruneL1SyncedBlocks()
}

func (m *Monitor) pruneL1SyncedBlocks() {
	latestL1Block, err := m.L1SyncedBlockModel.GetLatestL1Block()
	if err != nil {
		logx.Errorf("get oldest l1 block error, err: %s", err.Error())
		return
	}

	startL1BlockId := latestL1Block.ID - MaxRecordsToKeep
	for {
		if startL1BlockId <= 0 {
			return
		}

		l1Block, err := m.L1SyncedBlockModel.GetL1BlockById(startL1BlockId)
		if err != nil {
			logx.Errorf("get l1 block by id error, id=%d, err=%s", startL1BlockId, err.Error())
			return
		}

		logx.Infof("delete l1 synced block, id=%d", startL1BlockId)
		err = m.L1SyncedBlockModel.DeleteL1Block(l1Block)
		if err != nil {
			logx.Errorf("delete l1 block error, err=%s", err.Error())
			return
		}

		startL1BlockId = startL1BlockId - 1
	}
}
