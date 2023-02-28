package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	PriorityOperationMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_process",
		Help:      "Priority operation requestID metrics.",
	})
	PriorityOperationHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_process_height",
		Help:      "Priority operation height metrics.",
	})

	L2BlockMemoryHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_memory_height",
		Help:      "l2Block_memory_height metrics.",
	})

	L2BlockRedisHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_redis_height",
		Help:      "l2Block_memory_height metrics.",
	})

	L2BlockDbHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_db_height",
		Help:      "l2Block_memory_height metrics.",
	})

	AccountLatestVersionTreeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_account_latest_version",
		Help:      "Account latest version metrics.",
	})
	AccountRecentVersionTreeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_account_recent_version",
		Help:      "Account recent version metrics.",
	})
	NftTreeLatestVersionMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_nft_latest_version",
		Help:      "Nft latest version metrics.",
	})
	NftTreeRecentVersionMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_nft_recent_version",
		Help:      "Nft recent version metrics.",
	})

	CommitOperationMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "db_commit_time",
		Help:      "DB commit operation time",
	})
	ExecuteTxOperationMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_time",
		Help:      "execute txs operation time",
	})
	PendingTxNumMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "pending_tx",
		Help:      "number of pending tx",
	})
	UpdateAccountAssetTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_account_asset_tree_time",
		Help:      "updateAccountAssetTreeMetrics",
	})
	UpdateAccountTreeAndNftTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_account_tree_and_nft_tree_time",
		Help:      "updateAccountTreeAndNftTreeMetrics",
	})
	StateDBSyncOperationMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "state_sync_time",
		Help:      "stateDB sync operation time",
	})

	PreSaveBlockDataMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "pre_save_block_data_time",
		Help:      "pre save block data time",
	}, []string{"type"})

	SaveBlockDataMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "save_block_data_time",
		Help:      "save block data time",
	}, []string{"type"})

	FinalSaveBlockDataMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "final_save_block_data_time",
		Help:      "final save block data time",
	}, []string{"type"})

	ExecuteTxApply1TxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_apply_1_transaction_time",
		Help:      "execute txs apply 1 transaction operation time",
	})

	UpdatePoolTxsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_pool_txs_time",
		Help:      "update pool txs time",
	})

	AddCompressedBlockMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_compressed_block_time",
		Help:      "add compressed block time",
	})

	UpdateAccountMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_account_time",
		Help:      "update account time",
	})

	AddAccountHistoryMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_account_history_time",
		Help:      "add account history time",
	})

	AddTxDetailsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_tx_details_time",
		Help:      "add tx details time",
	})

	AddTxsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_txs_time",
		Help:      "add txs time",
	})

	DeletePoolTxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "delete_pool_tx_time",
		Help:      "delete pool tx time",
	})

	UpdateBlockMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_block_time",
		Help:      "update block time time",
	})

	SaveAccountsGoroutineMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "save_accounts_goroutine_time",
		Help:      "save accounts goroutine time",
	})

	GetPendingPoolTxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "get_pending_pool_tx_time",
		Help:      "get pending pool tx time",
	})

	UpdatePoolTxsProcessingMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_pool_txs_processing_time",
		Help:      "update pool txs processing time",
	})
	SyncAccountToRedisMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "sync_account_to_redis_time",
		Help:      "sync account to redis time",
	})
	GetPendingTxsToQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "get_pending_txs_to_queue_count",
		Help:      "get pending txs to queue count",
	})

	TxWorkerQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "tx_worker_queue_count",
		Help:      "tx worker queue count",
	})

	ExecuteTxMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "execute_tx_count",
		Help:      "execute tx count",
	})

	UpdateAssetTreeTxMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "update_account_asset_tree_tx_count",
		Help:      "update_account_asset_tree_tx_count",
	})
	UpdateAccountAndNftTreeTxMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "update_account_tree_and_nft_tree_tx_count",
		Help:      "update_account_tree_and_nft_tree_tx_count",
	})

	AccountAssetTreeQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_asset_tree_queue_count",
		Help:      "account asset tree queue count",
	})

	AccountAndNftTreeQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_tree_and_nft_tree_queue_count",
		Help:      "account tree and nft tree queue count",
	})

	AntsPoolGaugeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "ants_pool_count",
		Help:      "ants pool count",
	}, []string{"type"})

	L2BlockHeightMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2_block_height",
		Help:      "l2_Block_Height metrics.",
	})
	PoolTxL1ErrorCountMetics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "pool_tx_l1_error_count",
		Help:      "pool_tx_l1_error_count metrics.",
	})
	PoolTxL2ErrorCountMetics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "pool_tx_l2_error_count",
		Help:      "pool_tx_l2_error_count metrics.",
	})
)

// metrics
var (
	UpdateAssetTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_asset_smt",
		Help:      "update asset smt tree operation time",
	})

	CommitAccountTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_account_smt",
		Help:      "commit account smt tree operation time",
	})

	ExecuteTxPrepareMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_prepare_time",
		Help:      "execute txs prepare operation time",
	})

	ExecuteTxVerifyInputsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_verify_inputs_time",
		Help:      "execute txs verify inputs operation time",
	})

	ExecuteGenerateTxDetailsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_generate_tx_details_time",
		Help:      "execute txs generate tx details operation time",
	})

	ExecuteTxApplyTransactionMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_apply_transaction_time",
		Help:      "execute txs apply transaction operation time",
	})

	ExecuteTxGeneratePubDataMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_generate_pub_data_time",
		Help:      "execute txs generate pub data operation time",
	})
	ExecuteTxGetExecutedTxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_get_executed_tx_time",
		Help:      "execute txs get executed tx operation time",
	})

	AccountFromDbGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_from_db_time",
		Help:      "account from db time",
	})

	AccountGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_time",
		Help:      "account time",
	})

	VerifyGasGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "verifyGasGauge_time",
		Help:      "verifyGas time",
	})
	VerifySignature = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "verifySignature_time",
		Help:      "verifySignature time",
	})

	AccountTreeMultiSetGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "accountTreeMultiSetGauge_time",
		Help:      "accountTreeMultiSetGauge time",
	})
	GetAccountCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "get_account_counter",
		Help:      "get account counter",
	})

	GetAccountFromDbCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "get_account_from_db_counter",
		Help:      "get account from db counter",
	})
	AccountTreeTimeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "get_account_tree_time",
			Help:      "get_account_tree_time.",
		},
		[]string{"type"})
	NftTreeTimeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "zkbnb",
			Name:      "get_nft_tree_time",
			Help:      "get_nft_tree_time.",
		},
		[]string{"type"})
)

func InitCommitterMetrics() error {
	if err := prometheus.Register(PriorityOperationMetric); err != nil {
		return fmt.Errorf("prometheus.Register priorityOperationMetric error: %v", err)
	}
	if err := prometheus.Register(PriorityOperationHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register priorityOperationHeightMetric error: %v", err)
	}
	if err := prometheus.Register(L2BlockMemoryHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}
	if err := prometheus.Register(L2BlockRedisHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}
	if err := prometheus.Register(L2BlockDbHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}
	if err := prometheus.Register(AccountLatestVersionTreeMetric); err != nil {
		return fmt.Errorf("prometheus.Register AccountLatestVersionTreeMetric error: %v", err)
	}
	if err := prometheus.Register(AccountRecentVersionTreeMetric); err != nil {
		return fmt.Errorf("prometheus.Register AccountRecentVersionTreeMetric error: %v", err)
	}
	if err := prometheus.Register(NftTreeLatestVersionMetric); err != nil {
		return fmt.Errorf("prometheus.Register NftTreeLatestVersionMetric error: %v", err)
	}
	if err := prometheus.Register(NftTreeRecentVersionMetric); err != nil {
		return fmt.Errorf("prometheus.Register NftTreeRecentVersionMetric error: %v", err)
	}
	if err := prometheus.Register(CommitOperationMetics); err != nil {
		return fmt.Errorf("prometheus.Register commitOperationMetics error: %v", err)
	}
	if err := prometheus.Register(PendingTxNumMetrics); err != nil {
		return fmt.Errorf("prometheus.Register pendingTxNumMetrics error: %v", err)
	}
	if err := prometheus.Register(ExecuteTxOperationMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxOperationMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdateAccountAssetTreeMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountAssetTreeMetrics error: %v", err)
	}
	if err := prometheus.Register(StateDBSyncOperationMetics); err != nil {
		return fmt.Errorf("prometheus.Register stateDBSyncOperationMetics error: %v", err)
	}
	if err := prometheus.Register(PreSaveBlockDataMetrics); err != nil {
		return fmt.Errorf("prometheus.Register preSaveBlockDataMetrics error: %v", err)
	}
	if err := prometheus.Register(SaveBlockDataMetrics); err != nil {
		return fmt.Errorf("prometheus.Register saveBlockDataMetrics error: %v", err)
	}
	if err := prometheus.Register(FinalSaveBlockDataMetrics); err != nil {
		return fmt.Errorf("prometheus.Register finalSaveBlockDataMetrics error: %v", err)
	}
	if err := prometheus.Register(ExecuteTxApply1TxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxApply1TxMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdatePoolTxsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updatePoolTxsMetrics error: %v", err)
	}
	if err := prometheus.Register(AddCompressedBlockMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addCompressedBlockMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdateAccountMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountMetrics error: %v", err)
	}
	if err := prometheus.Register(AddAccountHistoryMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addAccountHistoryMetrics error: %v", err)
	}
	if err := prometheus.Register(AddTxDetailsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addTxDetailsMetrics error: %v", err)
	}
	if err := prometheus.Register(AddTxsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addTxsMetrics error: %v", err)
	}
	if err := prometheus.Register(DeletePoolTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register deletePoolTxMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdateBlockMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateBlockMetrics error: %v", err)
	}
	if err := prometheus.Register(SaveAccountsGoroutineMetrics); err != nil {
		return fmt.Errorf("prometheus.Register saveAccountsGoroutineMetrics error: %v", err)
	}
	if err := prometheus.Register(GetPendingPoolTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register getPendingPoolTxMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdatePoolTxsProcessingMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updatePoolTxsProcessingMetrics error: %v", err)
	}
	if err := prometheus.Register(SyncAccountToRedisMetrics); err != nil {
		return fmt.Errorf("prometheus.Register SyncAccountToRedisMetrics error: %v", err)
	}
	if err := prometheus.Register(GetPendingTxsToQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register getPendingTxsToQueueMetric error: %v", err)
	}
	if err := prometheus.Register(TxWorkerQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register txWorkerQueueMetric error: %v", err)
	}
	if err := prometheus.Register(ExecuteTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdateAssetTreeTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountAssetTreeTxMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdateAccountAndNftTreeTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountTreeAndNftTreeTxMetrics error: %v", err)
	}
	if err := prometheus.Register(UpdateAccountTreeAndNftTreeMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountTreeAndNftTreeMetrics error: %v", err)
	}
	if err := prometheus.Register(AccountAndNftTreeQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register accountTreeAndNftTreeQueueMetric error: %v", err)
	}
	if err := prometheus.Register(AccountAssetTreeQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register accountAssetTreeQueueMetric error: %v", err)
	}
	if err := prometheus.Register(AntsPoolGaugeMetric); err != nil {
		return fmt.Errorf("prometheus.Register antsPoolGaugeMetric error: %v", err)
	}
	if err := prometheus.Register(L2BlockHeightMetics); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockHeightMetics error: %v", err)
	}
	if err := prometheus.Register(PoolTxL1ErrorCountMetics); err != nil {
		return fmt.Errorf("prometheus.Register poolTxL1ErrorCountMetics error: %v", err)
	}
	if err := prometheus.Register(PoolTxL2ErrorCountMetics); err != nil {
		return fmt.Errorf("prometheus.Register poolTxL2ErrorCountMetics error: %v", err)
	}
	return nil
}

func InitBlockChainMetrics() error {
	if err := prometheus.Register(VerifyGasGauge); err != nil {
		return fmt.Errorf("prometheus.Register verifyGasGauge error: %v", err)
	}

	if err := prometheus.Register(VerifySignature); err != nil {
		return fmt.Errorf("prometheus.Register verifySignature error: %v", err)
	}

	if err := prometheus.Register(AccountTreeMultiSetGauge); err != nil {
		return fmt.Errorf("prometheus.Register accountTreeMultiSetGauge error: %v", err)
	}

	if err := prometheus.Register(AccountFromDbGauge); err != nil {
		return fmt.Errorf("prometheus.Register accountFromDbMetrics error: %v", err)
	}

	if err := prometheus.Register(GetAccountCounter); err != nil {
		return fmt.Errorf("prometheus.Register getAccountCounter error: %v", err)
	}

	if err := prometheus.Register(GetAccountFromDbCounter); err != nil {
		return fmt.Errorf("prometheus.Register getAccountFromDbCounter error: %v", err)
	}

	if err := prometheus.Register(AccountTreeTimeGauge); err != nil {
		return fmt.Errorf("prometheus.Register accountTreeTimeGauge error: %v", err)
	}

	if err := prometheus.Register(NftTreeTimeGauge); err != nil {
		return fmt.Errorf("prometheus.Register nftTreeTimeGauge error: %v", err)
	}

	if err := prometheus.Register(ExecuteTxPrepareMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxPrepareMetrics error: %v", err)
	}

	if err := prometheus.Register(ExecuteTxVerifyInputsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxVerifyInputsMetrics error: %v", err)
	}

	if err := prometheus.Register(ExecuteGenerateTxDetailsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeGenerateTxDetailsMetrics error: %v", err)
	}

	if err := prometheus.Register(ExecuteTxApplyTransactionMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxApplyTransactionMetrics error: %v", err)
	}

	if err := prometheus.Register(ExecuteTxGeneratePubDataMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxGeneratePubDataMetrics error: %v", err)
	}

	if err := prometheus.Register(ExecuteTxGetExecutedTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxGetExecutedTxMetrics error: %v", err)
	}

	if err := prometheus.Register(UpdateAssetTreeMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountTreeMetrics error: %v", err)
	}

	if err := prometheus.Register(CommitAccountTreeMetrics); err != nil {
		return fmt.Errorf("prometheus.Register commitAccountTreeMetrics error: %v", err)
	}
	return nil
}
