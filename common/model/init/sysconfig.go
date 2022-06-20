package main

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
)

const (
	// network rpc
	BSC_Test_Network_RPC   = "http://tf-dex-preview-validator-nlb-6fd109ac8b9d390a.elb.ap-northeast-1.amazonaws.com:8545"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"

	ZecreyLegendContractAddr = "0x3DD01F094a7D64c5B13b9ce0400Cf6cC61466C8A"
	GovernanceContractAddr   = "0x34fB7f0b5B69796248c745f8a2D2650ae1A014F7"
	VerifierContractAddr     = "0x194Ac316B82993fd0C482F6C3C1F69F261921490"
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
		//{
		//	Name:      sysconfigName.AssetGovernanceContract,
		//	Value:     AssetGovernanceContractAddr,
		//	ValueType: "string",
		//	Comment:   "Asset_Governance contract on BSC",
		//},

		// Verifier Contract
		//{
		//	Name:      sysconfigName.VerifierContract,
		//	Value:     VerifierContractAddr,
		//	ValueType: "string",
		//	Comment:   "Verifier contract on BSC",
		//},
		// network rpc
		{
			Name:      sysconfigName.BscTestNetworkRpc,
			Value:     BSC_Test_Network_RPC,
			ValueType: "string",
			Comment:   "BSC network rpc",
		},
		// TODO
		{
			Name:      sysconfigName.LocalTestNetworkRpc,
			Value:     Local_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Local network rpc",
		},
	}
}
