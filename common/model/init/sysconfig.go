package init

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
)

const (
	// network rpc
	BSC_Test_Network_RPC   = "https://data-seed-prebsc-1-s1.binance.org:8545/"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"

	ZecreyLegendContractAddr    = "0x0C17367D14760D2a54A3D140c9F2f1c2EdB81E7D"
	GovernanceContractAddr      = "0x4B7635b2A882F94cB4E50CDc073bA8630f1759A6"
	AssetGovernanceContractAddr = "0x3C4237AbEf419C7C76efAd854b7166F49C77F516"
	VerifierContractAddr        = "0x4EFCfA18c2cdf4661C028Df55F4911c7F82F253d"
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
			Name:      sysconfigName.TreasuryAccountIndex,
			Value:     "0",
			ValueType: "int",
			Comment:   "treasury index",
		},
		{
			Name:      sysconfigName.GasAccountIndex,
			Value:     "1",
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
