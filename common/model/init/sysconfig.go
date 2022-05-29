package init

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
)

const (
	// network rpc
	BSC_Test_Network_RPC   = "https://data-seed-prebsc-1-s1.binance.org:8545/"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"

	ZecreyLegendContractAddr    = "0xf1F49D2B2Fd889Bc3F3B565dEDD8CA32AFA4D0f2"
	GovernanceContractAddr      = "0x0FEA0F57a14eFCC8c4980a836AeFbD15b409C48C"
	AssetGovernanceContractAddr = "0xb3b641DD4cFb14724B9e254cb748C9e68F983B83"
	VerifierContractAddr        = "0x81d82a9f4bF0cE7e01782178fC87B221D1e01a73"
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
