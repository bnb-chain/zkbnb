/*
 * Copyright © 2021 Zecrey Protocol
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

package logic

import "github.com/bnb-chain/zkbas/common/util"

func GetUnprovedCryptoBlockStatus(blockNumber int64) (status int64) {
	for _, v := range UnProvedCryptoBlocks {
		if int64(v.BlockInfo.BlockNumber) == blockNumber {
			return v.Status
		}
	}
	return -1
}

func GetUnprovedCryptoBlock(mode int64) *CryptoBlockInfo {
	switch mode {
	case util.COO_MODE:
		for _, v := range UnProvedCryptoBlocks {
			if v.Status == PUBLISHED {
				return v
			}
		}
		break
	case util.COM_MODE:
		for _, v := range UnProvedCryptoBlocks {
			if v.Status <= RECEIVED {
				return v
			}
		}
		break
	default:
		return nil
	}
	return nil
}

func SetUnprovedCryptoBlockStatus(blockNumber int64, status int64) bool {
	for _, v := range UnProvedCryptoBlocks {
		if v.BlockInfo.BlockNumber == blockNumber {
			v.Status = status
			return true
		}
	}
	return false
}

func GetUnprovedCryptoBlockByBlockNumber(blockNumber int64) *CryptoBlockInfo {
	for _, v := range UnProvedCryptoBlocks {
		if v.BlockInfo.BlockNumber == blockNumber {
			return v
		}
	}
	return nil
}

func GetLatestUnprovedBlockHeight() (h int64) {
	if UnProvedCryptoBlocks == nil {
		return 0
	} else {
		return int64(UnProvedCryptoBlocks[len(UnProvedCryptoBlocks)-1].BlockInfo.BlockNumber)
	}

}
