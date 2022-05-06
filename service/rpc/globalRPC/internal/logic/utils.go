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

import (
	"errors"
	"fmt"
	"github.com/zecrey-labs/zecrey/common/commonAccount"
	"github.com/zecrey-labs/zecrey/common/commonAsset"
	"github.com/zecrey-labs/zecrey/common/commonTx"
	"github.com/zecrey-labs/zecrey/common/model/account"
	"github.com/zecrey-labs/zecrey/common/model/asset"
	"github.com/zecrey-labs/zecrey/common/model/l1amount"
	"github.com/zecrey-labs/zecrey/common/model/liquidityPair"
	"github.com/zecrey-labs/zecrey/common/model/mempool"
	"github.com/zecrey-labs/zecrey/common/utils"
	"github.com/zecrey-labs/zecrey/common/zcrypto/elgamal"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/globalmapHandler"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/txHandler"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetLatestSingleAccountAsset(svcCtx *svc.ServiceContext, accountIndex uint32, assetId uint32) (res *AccountSingleAsset, err error) {
	resAccount, err := svcCtx.AccountHistoryModel.GetAccountByAccountIndex(int64(accountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.GetLatestSingleAccountAsset] => [AccountModel.GetAccountByAccountIndex] %s. Invalid accountIndex %v",
			err.Error(), accountIndex)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	}

	// get latest account info by accountIndex and assetId
	globalKey := globalmapHandler.GetAccountAssetGlobalKey(uint32(resAccount.AccountIndex), assetId)
	ifExisted, globalValue := globalmapHandler.HandleGlobalMapGet(globalKey)

	// construct accountSingleAsset struct
	accountInfo := &AccountSingleAsset{
		AccountId:    resAccount.ID,
		AssetId:      assetId,
		AccountIndex: accountIndex,
		AccountName:  resAccount.AccountName,
		PublicKey:    resAccount.PublicKey,
	}

	if ifExisted {
		accountInfo.BalanceEnc = globalValue
	} else {
		// get accountAssetInfo by accountIndex and assetId
		resAccountSingleAsset, err := svcCtx.AssetModel.GetSingleAccountAsset(int64(accountIndex), int64(assetId))
		if err != nil {
			if err != asset.ErrNotFound {
				errInfo := fmt.Sprintf("[utils.GetLatestSingleAccountAsset] => [AssetModel.GetSingleAccountAsset] %s. Invalid accountIndex/assetId %v/%v",
					err.Error(), accountIndex, assetId)
				logx.Error(errInfo)
				return nil, errors.New(errInfo)
			} else {
				// init nil AccountSingleAsset
				resAccountSingleAsset = &asset.AccountAsset{
					AccountIndex: int64(accountIndex),
					AssetId:      int64(assetId),
					BalanceEnc:   elgamal.ZeroElgamalEncStr(),
				}
			}
		}
		// fetch all generalAssetType transaction including lock and deposit
		mempoolDetails, err := svcCtx.MempoolDetailModel.GetAccountAssetMempoolDetails(
			int64(accountInfo.AccountIndex),
			int64(accountInfo.AssetId),
			commonAsset.GeneralAssetType,
			commonTx.L2TxChainId,
		)
		accountInfo.BalanceEnc = resAccountSingleAsset.BalanceEnc
		if err != nil {
			if err != mempool.ErrNotFound {
				errInfo := fmt.Sprintf("[sendtxlogic.GetLatestSingleAccountAsset] => [MempoolDetailModel.GetAccountAssetMempoolDetails] %s",
					err.Error())
				logx.Error(errInfo)
				return nil, errors.New(errInfo)
			}
		}
		finalBalance, err := globalmapHandler.UpdateSingleAssetGlobalMapByMempoolDetails(mempoolDetails, &globalmapHandler.GlobalAssetInfo{
			AccountIndex:   int64(accountInfo.AccountIndex),
			AssetId:        int64(accountInfo.AssetId),
			AssetType:      commonAsset.GeneralAssetType,
			ChainId:        commonTx.L2TxChainId,
			BaseBalanceEnc: resAccountSingleAsset.BalanceEnc,
		})
		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.GetLatestSingleAccountAsset] => [UpdateSingleAssetGlobalMapByMempoolDetails] %s",
				err.Error())
			logx.Error(errInfo)
			return nil, errors.New(errInfo)
		}
		accountInfo.BalanceEnc = finalBalance
	}
	return accountInfo, err
}

func GetLatestPoolInfo(svcCtx *svc.ServiceContext, pairIndex uint32) (
	pairInfo *liquidityPair.LiquidityPair,
	poolAccountInfo *account.AccountHistory,
	poolLiquidity *asset.AccountLiquidity,
	poolAssetAAmount,
	poolAssetBAmount uint64,
	err error) {

	pairInfo, err = svcCtx.LiquidityPairModel.GetLiquidityPairByIndex(pairIndex)
	if err != nil {
		errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [LiquidityPairModel.GetLiquidityPairByIndex]: %s. Invalid PairIndex: %v",
			err.Error(), pairIndex)
		logx.Error(errInfo)
		return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
	}

	poolAccountInfo, err = svcCtx.AccountHistoryModel.GetAccountByAccountIndex(commonAccount.PoolAccountIndex)
	if err != nil {
		errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [AccountModel.GetAccountByAccountIndex]: %s",
			err.Error())
		logx.Error(errInfo)
		return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
	}

	globalKey := globalmapHandler.GetPoolLiquidityGlobalKey(pairIndex)
	ifExisted, globalValue := globalmapHandler.HandleGlobalMapGet(globalKey)

	if ifExisted {
		latestPairInfo, err := utils.ParsePairInfo(globalValue)
		if err != nil {
			errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [ParsePairInfo]: %s. nPairInfo: %s",
				err.Error(), globalValue)
			logx.Error(errInfo)
			return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
		}
		poolLiquidity = &asset.AccountLiquidity{
			PairIndex: int64(pairIndex),
			AssetA:    latestPairInfo.PoolA.DeltaAmount,
			AssetAR:   latestPairInfo.PoolA.DeltaR.String(),
			AssetB:    latestPairInfo.PoolB.DeltaAmount,
			AssetBR:   latestPairInfo.PoolB.DeltaR.String(),
			LpEnc:     elgamal.ZeroElgamalEncStr(),
		}
		return pairInfo, poolAccountInfo, poolLiquidity, uint64(poolLiquidity.AssetA), uint64(poolLiquidity.AssetB), nil
	} else {
		poolLiquidity, err = svcCtx.LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex(commonAccount.PoolAccountIndex, pairIndex)
		if err != nil {
			// check if the err is ErrNotFound
			if err == asset.ErrNotFound {
				poolLiquidity = &asset.AccountLiquidity{
					AccountIndex: commonAccount.PoolAccountIndex,
					PairIndex:    int64(pairIndex),
					AssetA:       0,
					AssetB:       0,
					AssetAR:      "0",
					AssetBR:      "0",
					LpEnc:        elgamal.ZeroElgamalEnc().String(),
				}
				err = svcCtx.LiquidityAssetModel.CreateAccountLiquidity(poolLiquidity)
				if err != nil {
					errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex]: %s. Invalid PairIndex: %v",
						err.Error(), pairIndex)
					logx.Error(errInfo)
					return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
				}
			} else {
				errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex]: %s. Invalid PairIndex: %v",
					err.Error(), pairIndex)
				logx.Error(errInfo)
				return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
			}
		}

		// fetch all poolAccountIndex related transaction
		mempoolDetails, err := svcCtx.MempoolDetailModel.GetAccountAssetMempoolDetails(
			int64(commonAccount.PoolAccountIndex),
			int64(pairIndex),
			commonAsset.LiquidityAssetType,
			commonTx.L2TxChainId,
		)
		if err != nil {
			if err != mempool.ErrNotFound {
				errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [MempoolDetailModel.GetAccountAssetMempoolDetails] %s",
					err.Error())
				logx.Error(errInfo)
				return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
			}
		}
		pInfo, err := utils.ConstructPairInfo(poolLiquidity.AssetA, poolLiquidity.AssetB, poolLiquidity.AssetAR, poolLiquidity.AssetBR)
		if err != nil {
			errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [ConstructPairInfo]: %s. Invalid pairInfo: %v",
				err.Error(), poolLiquidity)
			logx.Error(errInfo)
			return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
		}
		finalBalance, err := globalmapHandler.UpdateSingleAssetGlobalMapByMempoolDetails(mempoolDetails, &globalmapHandler.GlobalAssetInfo{
			AccountIndex:   int64(commonAccount.PoolAccountIndex),
			AssetId:        int64(pairIndex),
			AssetType:      commonAsset.LiquidityAssetType,
			ChainId:        commonTx.L2TxChainId,
			BaseBalanceEnc: pInfo.String(),
		})
		if err != nil {
			errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [UpdateSingleAssetGlobalMapByMempoolDetails] %s",
				err.Error())
			logx.Error(errInfo)
			return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
		}

		latestPairInfo, err := utils.ParsePairInfo(finalBalance)
		if err != nil {
			errInfo := fmt.Sprintf("[GetLatestPoolInfo] => [ParsePairInfo]: %s. nPairInfo: %s",
				err.Error(), finalBalance)
			logx.Error(errInfo)
			return pairInfo, poolAccountInfo, poolLiquidity, poolAssetAAmount, poolAssetBAmount, errors.New(errInfo)
		}
		poolLiquidity.AssetA = latestPairInfo.PoolA.DeltaAmount
		poolLiquidity.AssetAR = latestPairInfo.PoolA.DeltaR.String()
		poolLiquidity.AssetB = latestPairInfo.PoolB.DeltaAmount
		poolLiquidity.AssetBR = latestPairInfo.PoolB.DeltaR.String()

		return pairInfo, poolAccountInfo, poolLiquidity, uint64(poolLiquidity.AssetA), uint64(poolLiquidity.AssetB), nil
	}
}

func GetLatestLockedAsset(
	svcCtx *svc.ServiceContext,
	accountIndex uint32,
	chainId uint8,
	assetId uint32,
) (res *AccountSingleLockedAsset, err error) {
	resAccount, err := svcCtx.AccountHistoryModel.GetAccountByAccountIndex(int64(accountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [AccountModel.GetAccountByAccountIndex] %s. Invalid accountIndex %v",
			err.Error(), accountIndex)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	}

	// get latest account info by accountIndex and assetId
	globalKey := globalmapHandler.GetAccountLockAssetGlobalKey(accountIndex, chainId, assetId)
	ifExisted, globalValue := globalmapHandler.HandleGlobalMapGet(globalKey)

	// construct accountSingleAsset struct
	accountInfo := &AccountSingleLockedAsset{
		AccountId:    resAccount.ID,
		AccountIndex: uint32(resAccount.AccountIndex),
		AccountName:  resAccount.AccountName,
		PublicKey:    resAccount.PublicKey,
		ChainId:      chainId,
		AssetId:      assetId,
	}

	if ifExisted {
		nBalanceInt, err := utils.StringToUint64(globalValue)
		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [utils.StringToUint64] %s",
				err.Error())
			logx.Error(errInfo)
			return nil, errors.New(errInfo)
		}
		accountInfo.LockedAmount = nBalanceInt
	} else {
		// get accountLockedAssetInfo by accountIndex and assetId and chainId
		resAccountSingleLockedAsset, err := svcCtx.LockAssetModel.GetAccountAssetLockByIndexAndChainIdAndAssetId(int64(accountIndex), int64(chainId), int64(assetId))
		if err != nil {
			if err != asset.ErrNotFound {
				errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [LockAssetModel.GetAccountAssetLockByIndexAndChainIdAndAssetId] %s. Invalid accountIndex/assetId %v/%v",
					err.Error(), accountIndex, assetId)
				logx.Error(errInfo)
				return nil, errors.New(errInfo)
			} else {
				resAccountSingleLockedAsset = &asset.AccountAssetLock{
					ChainId:      int64(chainId),
					AccountIndex: int64(accountIndex),
					AssetId:      int64(assetId),
					LockedAmount: 0,
				}
			}
		}
		// fetch mempool detail including lock and unlock transaction
		// todo test after creating mempool details
		mempoolDetails, err := svcCtx.MempoolDetailModel.GetAccountAssetMempoolDetails(
			int64(accountIndex),
			int64(assetId),
			commonAsset.LockedAssetType,
			int64(chainId),
		)
		if err != nil {
			if err != mempool.ErrNotFound {
				errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [MempoolDetailModel.GetAccountAssetMempoolDetails] %s",
					err.Error())
				logx.Error(errInfo)
				return nil, errors.New(errInfo)
			}
		}
		oBalance := utils.Uint64ToString(uint64(resAccountSingleLockedAsset.LockedAmount))
		finalBalance, err := globalmapHandler.UpdateSingleAssetGlobalMapByMempoolDetails(mempoolDetails, &globalmapHandler.GlobalAssetInfo{
			AccountIndex:   resAccountSingleLockedAsset.AccountIndex,
			AssetId:        resAccountSingleLockedAsset.AssetId,
			AssetType:      commonAsset.LockedAssetType,
			ChainId:        int64(chainId),
			BaseBalanceEnc: oBalance,
		})
		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [UpdateSingleAssetGlobalMapByMempoolDetails] %s",
				err.Error())
			logx.Error(errInfo)
			return nil, errors.New(errInfo)
		}
		finalBalanceInt64, err := utils.StringToUint64(finalBalance)
		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [utils.StringToUint64] %s",
				err.Error())
			logx.Error(errInfo)
			return nil, errors.New(errInfo)
		}
		accountInfo.LockedAmount = finalBalanceInt64
	}
	return accountInfo, nil
}

func GetLatestAccountLpInfo(svcCtx *svc.ServiceContext, accountIndex uint32, pairIndex uint32) (
	pairInfo *liquidityPair.LiquidityPair,
	accountLiquidity *asset.AccountLiquidity,
	accountLpEnc string,
	err error) {

	// todo pairInfo cache
	pairInfo, err = svcCtx.LiquidityPairModel.GetLiquidityPairByIndex(pairIndex)
	if err != nil {
		errInfo := fmt.Sprintf("[GetLatestAccountLpInfo] => [LiquidityPairModel.GetLiquidityPairByIndex]: %s. Invalid PairIndex: %v",
			err.Error(), pairIndex)
		logx.Error(errInfo)
		return pairInfo, accountLiquidity, accountLpEnc, errors.New(errInfo)
	}

	accountLiquidity, err = svcCtx.LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex(accountIndex, pairIndex)
	if err != nil {
		// if account liquidity pair doesn't exist then create new records for account Liquidity
		if err == asset.ErrNotFound {
			accountLiquidity = &asset.AccountLiquidity{
				AccountIndex: int64(accountIndex),
				PairIndex:    int64(pairIndex),
				AssetA:       0,
				AssetB:       0,
				AssetAR:      "0",
				AssetBR:      "0",
				LpEnc:        elgamal.ZeroElgamalEnc().String(),
			}
			// create accountLiquidity for accountLiquidityAsset query
			err = svcCtx.LiquidityAssetModel.CreateAccountLiquidity(accountLiquidity)
			if err != nil {
				errInfo := fmt.Sprintf("[GetLatestAccountLpInfo] => [LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex][CreateAccountLiquidity]: %s. accountLiquidity: %v",
					err.Error(), accountLiquidity)
				logx.Error(errInfo)
				return pairInfo, accountLiquidity, accountLpEnc, errors.New(errInfo)
			}
		} else {
			errInfo := fmt.Sprintf("[GetLatestAccountLpInfo] => [LiquidityAssetModel.GetLiquidityByAccountIndexAndPairIndex]: %s. Invalid AccountIndex/PairIndex: %v / %v",
				err.Error(), accountIndex, pairIndex)
			logx.Error(errInfo)
			return pairInfo, accountLiquidity, accountLpEnc, errors.New(errInfo)
		}
	}

	globalKey := globalmapHandler.GetAccountLPGlobalKey(accountIndex, pairIndex)
	ifExisted, globalValue := globalmapHandler.HandleGlobalMapGet(globalKey)

	if ifExisted {
		// todo accountLiquidity.LpEnc = globalValue
		return pairInfo, accountLiquidity, globalValue, nil
	} else {
		mempoolDetails, err := svcCtx.MempoolDetailModel.GetAccountAssetMempoolDetails(
			accountLiquidity.AccountIndex,
			accountLiquidity.PairIndex,
			commonAsset.LiquidityLpAssetType,
			commonTx.L2TxChainId,
		)
		if err != nil {
			if err != mempool.ErrNotFound {
				errInfo := fmt.Sprintf("[sendtxlogic.GetLatestAccountLpInfo] => [MempoolDetailModel.GetAccountAssetMempoolDetails] %s",
					err.Error())
				logx.Error(errInfo)
				return pairInfo, accountLiquidity, accountLiquidity.LpEnc, errors.New(errInfo)
			}
		}
		finalBalance, err := globalmapHandler.UpdateSingleAssetGlobalMapByMempoolDetails(mempoolDetails, &globalmapHandler.GlobalAssetInfo{
			AccountIndex:   accountLiquidity.AccountIndex,
			AssetId:        accountLiquidity.PairIndex,
			AssetType:      commonAsset.LiquidityLpAssetType,
			ChainId:        commonTx.L2TxChainId,
			BaseBalanceEnc: accountLiquidity.LpEnc,
		})
		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.GetLatestSingleAccountAsset] => [UpdateSingleAssetGlobalMapByMempoolDetails] %s",
				err.Error())
			logx.Error(errInfo)
			return pairInfo, accountLiquidity, accountLiquidity.LpEnc, errors.New(errInfo)
		}
		accountLiquidity.LpEnc = finalBalance
		return pairInfo, accountLiquidity, accountLiquidity.LpEnc, nil
	}

}

func GetLatestL1Amount(svcCtx *svc.ServiceContext, chainId uint32, assetId uint32) (
	l1FinalAmount int64,
	err error,
) {
	globalKey := globalmapHandler.GetL1AmountGlobalKey(uint8(chainId), assetId)
	ifExisted, globalValue := globalmapHandler.HandleGlobalMapGet(globalKey)
	if ifExisted {
		l1FinalAmount, err = utils.StringToInt64(globalValue)
		if err != nil {
			logx.Errorf("[GetLatestL1Amount] StringToInt64 error : %s", err.Error())
			return 0, err
		} else {
			return l1FinalAmount, nil
		}
	} else {
		// init l1Amount
		l1FinalAmount = 0

		l1AssetAmount, err := svcCtx.L1AmountModel.GetL1AmountByChainIdAndAssetId(uint8(chainId), assetId)
		if err != nil {
			if err != l1amount.ErrNotFound {
				errInfo := fmt.Sprintf("[GetLatestL1Amount] => [L1AmountModel.GetLatestL1AmountInfo] %s",
					err.Error())
				logx.Info(errInfo)
				return 0, errors.New(errInfo)
			} else {
				errInfo := fmt.Sprintf("[GetLatestL1Amount] => [L1AmountModel.GetLatestL1AmountInfo] %s",
					err.Error())
				logx.Error(errInfo)
			}
		}

		l1FinalAmount = l1AssetAmount

		mempoolTxs, err := svcCtx.MempoolModel.GetL1MempoolTx(int64(chainId), int64(assetId))
		if err != nil {
			if err != mempool.ErrNotFound {
				errInfo := fmt.Sprintf("[GetLatestL1Amount] => [L1AmountModel.GetL1MempoolTx] %s",
					err.Error())
				logx.Info(errInfo)
				return 0, errors.New(errInfo)
			} else {
				errInfo := fmt.Sprintf("[GetLatestL1Amount] => [L1AmountModel.GetL1MempoolTx] %s",
					err.Error())
				logx.Error(errInfo)
			}
		}
		for _, mempoolTx := range mempoolTxs {
			switch mempoolTx.TxType {
			case commonTx.TxTypeDeposit:
				l1FinalAmount += mempoolTx.TxAmount
				break
			case commonTx.TxTypeLock:
				l1FinalAmount += mempoolTx.TxAmount
				break
			case commonTx.TxTypeWithdraw:
				l1FinalAmount -= mempoolTx.TxAmount
				break
			default:
				errInfo := fmt.Sprintf("[GetLatestL1Amount] txType error: %v", mempoolTx.TxType)
				logx.Error(errInfo)
				return 0, errors.New(errInfo)
			}
		}

		// update globalMap
		l1FinalAmountString := utils.Int64ToString(l1FinalAmount)
		globalmapHandler.HandleGlobalMapUpdate(globalKey, l1FinalAmountString)
	}

	return l1FinalAmount, err
}

func GetLatestLockedAssetList(
	svcCtx *svc.ServiceContext,
	accountIndex uint32,
) (res []*AccountSingleLockedAsset, err error) {
	resAccount, err := svcCtx.AccountHistoryModel.GetAccountByAccountIndex(int64(accountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [AccountModel.GetAccountByAccountIndex] %s. Invalid accountIndex %v",
			err.Error(), accountIndex)
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	}

	l1AssetList, err := svcCtx.L1AssetModel.GetAssets()
	if err != nil {
		errInfo := fmt.Sprintf("[GetLatestLockedAssetList] => [L1AssetModel.GetAssets] %s", err.Error())
		logx.Error(errInfo)
		return nil, errors.New(errInfo)
	}

	for _, v := range l1AssetList {
		chainId := uint8(v.ChainId)
		assetId := uint32(v.AssetId)

		// get latest account info by accountIndex and assetId
		globalKey := globalmapHandler.GetAccountLockAssetGlobalKey(accountIndex, chainId, assetId)
		ifExisted, globalValue := globalmapHandler.HandleGlobalMapGet(globalKey)

		// construct accountSingleAsset struct
		accountInfo := &AccountSingleLockedAsset{
			AccountId:    resAccount.ID,
			AccountIndex: uint32(resAccount.AccountIndex),
			AccountName:  resAccount.AccountName,
			PublicKey:    resAccount.PublicKey,
			ChainId:      chainId,
			AssetId:      assetId,
			AssetName:    v.AssetName,
		}

		if ifExisted {
			nBalanceInt, err := utils.StringToUint64(globalValue)
			if err != nil {
				errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [utils.StringToUint64] %s",
					err.Error())
				logx.Error(errInfo)
				return nil, errors.New(errInfo)
			}
			accountInfo.LockedAmount = nBalanceInt
		} else {
			// get accountLockedAssetInfo by accountIndex and assetId and chainId
			resAccountSingleLockedAsset, err := svcCtx.LockAssetModel.GetAccountAssetLockByIndexAndChainIdAndAssetId(int64(accountIndex), int64(chainId), int64(assetId))
			if err != nil {
				if err != asset.ErrNotFound {
					errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [LockAssetModel.GetAccountAssetLockByIndexAndChainIdAndAssetId] %s. Invalid accountIndex/assetId %v/%v",
						err.Error(), accountIndex, assetId)
					logx.Error(errInfo)
					return nil, errors.New(errInfo)
				} else {
					resAccountSingleLockedAsset = &asset.AccountAssetLock{
						ChainId:      int64(chainId),
						AccountIndex: int64(accountIndex),
						AssetId:      int64(assetId),
						LockedAmount: 0,
					}
				}
			}
			// fetch mempool detail including lock and unlock transaction
			// todo test after creating mempool details
			mempoolDetails, err := svcCtx.MempoolDetailModel.GetAccountAssetMempoolDetails(
				int64(accountIndex),
				int64(assetId),
				commonAsset.LockedAssetType,
				int64(chainId),
			)
			if err != nil {
				if err != mempool.ErrNotFound {
					errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [MempoolDetailModel.GetAccountAssetMempoolDetails] %s",
						err.Error())
					logx.Error(errInfo)
					return nil, errors.New(errInfo)
				}
			}
			oBalance := utils.Uint64ToString(uint64(resAccountSingleLockedAsset.LockedAmount))
			finalBalance, err := globalmapHandler.UpdateSingleAssetGlobalMapByMempoolDetails(mempoolDetails, &globalmapHandler.GlobalAssetInfo{
				AccountIndex:   resAccountSingleLockedAsset.AccountIndex,
				AssetId:        resAccountSingleLockedAsset.AssetId,
				AssetType:      commonAsset.LockedAssetType,
				ChainId:        int64(chainId),
				BaseBalanceEnc: oBalance,
			})
			if err != nil {
				errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [UpdateSingleAssetGlobalMapByMempoolDetails] %s",
					err.Error())
				logx.Error(errInfo)
				return nil, errors.New(errInfo)
			}
			finalBalanceInt64, err := utils.StringToUint64(finalBalance)
			if err != nil {
				errInfo := fmt.Sprintf("[sendtxlogic.GetLatestLockedAsset] => [utils.StringToUint64] %s",
					err.Error())
				logx.Error(errInfo)
				return nil, errors.New(errInfo)
			}
			accountInfo.LockedAmount = finalBalanceInt64
		}
		res = append(res, accountInfo)
	}

	return res, nil
}


func GetTxTypeArray(txType uint) ([]uint8, error) {
	switch txType {
	case L2TransferType:
		return []uint8{txHandler.TxTypeTransfer}, nil
	case LiquidityType:
		return []uint8{txHandler.TxTypeAddLiquidity, txHandler.TxTypeRemoveLiquidity}, nil
	case L2SwapType:
		return []uint8{txHandler.TxTypeSwap}, nil
	case WithdrawAssetsType:
		return []uint8{txHandler.TxTypeWithdraw}, nil
	case EncryptAssetsType:
		return []uint8{txHandler.TxTypeLock, txHandler.TxTypeUnLock}, nil
	default:
		errInfo := fmt.Sprintf("[GetTxTypeArray] txType error: %v", txType)
		logx.Error(errInfo)
		return []uint8{}, errors.New(errInfo)
	}
}
