package init

import (
	"github.com/zecrey-labs/zecrey-core/common/general/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
)

const (
	// network rpc
	BSC_Test_Network_RPC   = "https://data-seed-prebsc-1-s1.binance.org:8545/"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"

	ZecreyLegendContractAddr    = "0xD27624F18D423b990A7859A3d72916A3DD783EEE"
	GovernanceContractAddr      = "0xc72610F6c9296377AFbb4A58D51b0B079775a9C3"
	AssetGovernanceContractAddr = "0x24EF26Fb0c08c1c4f30ec736E14F9d53204f4de3"
	VerifierContractAddr        = "0xc26F98e338c8dE447fF6aC519C70560a4057B5CC"
)

func initSysConfig() []*sysconfig.Sysconfig {
	return []*sysconfig.Sysconfig{
		{
			Name:      sysconfigName.SysGasFee,
			Value:     "1",
			ValueType: "float",
			Comment:   "based on ETH",
		},
		{
			Name:      sysconfigName.MaxAssetId,
			Value:     "9",
			ValueType: "int",
			Comment:   "max number of asset id",
		},
		{
			Name:      sysconfigName.PoolAccountIndex,
			Value:     "0",
			ValueType: "int",
			Comment:   "pool index",
		},
		{
			Name:      sysconfigName.TreasuryAccountIndex,
			Value:     "1",
			ValueType: "int",
			Comment:   "treasury index",
		},
		{
			Name:      sysconfigName.GasAccountIndex,
			Value:     "2",
			ValueType: "int",
			Comment:   "gas index",
		},
		{
			Name:      sysconfigName.ZecreyLegendContract,
			Value:     ZecreyLegendContractAddr,
			ValueType: "string",
			Comment:   "Zecrey contract on BSC",
		},
		// Governance Contract
		{
			Name:      sysconfigName.GovernanceContract,
			Value:     GovernanceContractAddr,
			ValueType: "string",
			Comment:   "Governance contract on BSC",
		},

		// Asset_Governance Contract
		{
			Name:      sysconfigName.AssetGovernanceContract,
			Value:     AssetGovernanceContractAddr,
			ValueType: "string",
			Comment:   "Asset_Governance contract on BSC",
		},

		// Verifier Contract
		{
			Name:      sysconfigName.VerifierContract,
			Value:     VerifierContractAddr,
			ValueType: "string",
			Comment:   "Verifier contract on BSC",
		},
		// network rpc
		{
			Name:      sysconfigName.BscTestNetworkRpc,
			Value:     BSC_Test_Network_RPC,
			ValueType: "string",
			Comment:   "BSC network rpc",
		},
		// TODO
		{
			Name:      "Local_Test_Network_RPC",
			Value:     Local_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Local network rpc",
		},
	}
}
