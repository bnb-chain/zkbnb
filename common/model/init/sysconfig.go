package init

import (
	"github.com/zecrey-labs/zecrey-core/common/general/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
)

const (
	// network rpc
	BSC_Test_Network_RPC   = "https://data-seed-prebsc-1-s1.binance.org:8545/"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"

	ZecreyLegendContractAddr    = "0x700A9288d64e9994eF29CA09BE857A6d831850B4"
	GovernanceContractAddr      = "0x21c8ec99655bCaF34E09c1EdC25E0b12838A0620"
	AssetGovernanceContractAddr = "0x963753634DEa4fBC85DB8Ea0dA08eF3ae2b30dcF"
	VerifierContractAddr        = "0x37dC1a2aA1C21668b10d2447C62c7A0A24d18C69"
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
