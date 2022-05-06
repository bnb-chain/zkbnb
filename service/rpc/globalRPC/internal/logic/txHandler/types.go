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

package txHandler

type MempoolTx struct {
	TxHash       string
	TxType       uint32
	TxFee        uint32
	TxFeeAssetId uint32
	TxAssetAId   uint32
	TxAssetBId   int64
	TxAmount     int64
	TxTo         string
	TxProof      string
	ExtraInfo    string
	TxDetail     []*MempoolTxDetail
	Memo         string
}

type MempoolTxDetail struct {
	Nonce             int64
	AccountIndex      int64
	AccountName       string
	AccountBalanceEnc string
	AccountDeltaEnc   string
}

type Tx struct {
	TxHash       string
	TxType       uint32
	TxFee        uint32
	TxFeeAssetId uint32
	TxStatus     uint32
	BlockHeight  int64
	BlockId      uint32
	BlockStatus  uint32
	AccountRoot  string
	TxAssetAId   uint32
	TxAssetBId   int64
	TxAmount     int64
	TxTo         string
	TxProof      string
	ExtraInfo    string
	TxDetail     []*TxDetail
	Memo         string
}

type TxDetail struct {
	Nonce             int64
	AccountIndex      int64
	AccountName       string
	AccountBalanceEnc string
	AccountDeltaEnc   string
}
