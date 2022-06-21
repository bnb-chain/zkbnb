package main

import (
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
)

const (
	// network rpc
	//BSC_Test_Network_RPC   = "http://tf-dex-preview-validator-nlb-6fd109ac8b9d390a.elb.ap-northeast-1.amazonaws.com:8545"
	BSC_Test_Network_RPC   = "https://data-seed-prebsc-2-s1.binance.org:8545/"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"

	ZecreyLegendContractAddr = "0x915Bf51F6C9Fff38730C4557e79E1dEC4C01355a"
	GovernanceContractAddr   = "0xb081E0e1f243A32B3310eCdA15Ecb5B8189a8851"
	VerifierContractAddr     = "0x49113B180cAF7fb927006334a7c17eed6Dd92f70"
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
			Name:      "LocalTestNetworkRpc",
			Value:     Local_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Local network rpc",
		},
	}
}
