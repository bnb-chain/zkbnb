package init

import (
	"github.com/zecrey-labs/zecrey/common/model/sysconfig"
)

const (
	Main_Chain_Id = "Main_Chain_Id"
	Sys_Gas_Fee   = "Sys_Gas_Fee"
	Max_Chain_Id  = "Max_Chain_Id"
	Max_Asset_Id  = "Max_Asset_Id"
	// Zecrey Contract
	Zecrey_Contract_Ethereum    = "Zecrey_Contract_Ethereum"
	Zecrey_Contract_Polygon     = "Zecrey_Contract_Polygon"
	Zecrey_Contract_NEAR_Aurora = "Zecrey_Contract_NEAR_Aurora"
	Zecrey_Contract_Avalanche   = "Zecrey_Contract_Avalanche"
	Zecrey_Contract_BSC         = "Zecrey_Contract_BSC"
	// Governance Contract
	Governance_Contract_Ethereum    = "Governance_Contract_Ethereum"
	Governance_Contract_Polygon     = "Governance_Contract_Polygon"
	Governance_Contract_NEAR_Aurora = "Governance_Contract_NEAR_Aurora"
	Governance_Contract_Avalanche   = "Governance_Contract_Avalanche"
	Governance_Contract_BSC         = "Governance_Contract_BSC"
	// Asset Governance Contract
	Asset_Governance_Contract_Ethereum    = "Asset_Governance_Contract_Ethereum"
	Asset_Governance_Contract_Polygon     = "Asset_Governance_Contract_Polygon"
	Asset_Governance_Contract_NEAR_Aurora = "Asset_Governance_Contract_NEAR_Aurora"
	Asset_Governance_Contract_Avalanche   = "Asset_Governance_Contract_Avalanche"
	Asset_Governance_Contract_BSC         = "Asset_Governance_Contract_BSC"
	// Verifier Contract
	Verifier_Contract_Ethereum    = "Verifier_Contract_Ethereum"
	Verifier_Contract_Polygon     = "Verifier_Contract_Polygon"
	Verifier_Contract_NEAR_Aurora = "Verifier_Contract_NEAR_Aurora"
	Verifier_Contract_Avalanche   = "Verifier_Contract_Avalanche"
	Verifier_Contract_BSC         = "Verifier_Contract_BSC"
	// network rpc
	Ethereum_Test_Network_RPC  = "https://rinkeby.infura.io/v3/787bc04c3f044d77a538d519ef26b53e"
	Polygon_Test_Network_RPC   = "https://polygon-mumbai.g.alchemy.com/v2/O2mVU_nX6p-nnrTFKqASBQi74hsxCsro"
	Aurora_Test_Network_RPC    = "https://testnet.aurora.dev"
	Avalanche_Test_Network_RPC = "https://api.avax-test.network/ext/bc/C/rpc"
	BSC_Test_Network_RPC       = "https://data-seed-prebsc-1-s1.binance.org:8545/"
)

func initSysConfig() []*sysconfig.Sysconfig {
	return []*sysconfig.Sysconfig{
		{
			Name:      Main_Chain_Id,
			Value:     "0",
			ValueType: "int",
			Comment:   "main chain id",
		},
		{
			Name:      Sys_Gas_Fee,
			Value:     "1",
			ValueType: "float",
			Comment:   "based on ETH",
		},
		{
			Name:      Max_Chain_Id,
			Value:     "4",
			ValueType: "int",
			Comment:   "max number of chain id",
		},
		{
			Name:      Max_Asset_Id,
			Value:     "9",
			ValueType: "int",
			Comment:   "max number of asset id",
		},
		// Zecrey Contract
		{
			Name:      Zecrey_Contract_Ethereum,
			Value:     "0x190387Ae3AC96DC4fbAD820EC5819A8A5289E8D5",
			ValueType: "string",
			Comment:   "Zecrey contract on Ethereum",
		},
		{
			Name:      Zecrey_Contract_Polygon,
			Value:     "0x276A35a7E9D8506EDC6Bbf0aA0117dC0358B2748",
			ValueType: "string",
			Comment:   "Zecrey contract on Polygon",
		},
		{
			Name:      Zecrey_Contract_NEAR_Aurora,
			Value:     "0x18a3612BCBB6df5B3bD30DBaeC5ad321D3d1B4F9",
			ValueType: "string",
			Comment:   "Zecrey contract on Aurora",
		},
		{
			Name:      Zecrey_Contract_Avalanche,
			Value:     "0xe3867cDDf60b7bf2bE2FDb9919248e2186A66dFf",
			ValueType: "string",
			Comment:   "Zecrey contract on Avalanche",
		},
		{
			Name:      Zecrey_Contract_BSC,
			Value:     "0xeA1AD8BDc4281BB0bD892147b3c812C253895669",
			ValueType: "string",
			Comment:   "Zecrey contract on BSC",
		},
		// Governance Contract
		{
			Name:      Governance_Contract_Ethereum,
			Value:     "0xDbcbe97ef4D07aC4093517320148602b4F7b49D0",
			ValueType: "string",
			Comment:   "Governance contract on Ethereum",
		},
		{
			Name:      Governance_Contract_Polygon,
			Value:     "0x2C3697810eEea3cbaa5565816227b746C61DA514",
			ValueType: "string",
			Comment:   "Governance contract on Polygon",
		},
		{
			Name:      Governance_Contract_NEAR_Aurora,
			Value:     "0x3D1E02634C6c7B2e0640849E6584D332D1B90AD1",
			ValueType: "string",
			Comment:   "Governance contract on Aurora",
		},
		{
			Name:      Governance_Contract_Avalanche,
			Value:     "0x0fD65e8db55fEc604dFf52d80F9d7bcd3fee6C8F",
			ValueType: "string",
			Comment:   "Governance contract on Avalanche",
		},
		{
			Name:      Governance_Contract_BSC,
			Value:     "0x97E782f13402783090C0017CB89b5Fc71f608F06",
			ValueType: "string",
			Comment:   "Governance contract on BSC",
		},

		// Asset_Governance Contract
		{
			Name:      Asset_Governance_Contract_Ethereum,
			Value:     "0xA6b20426fec0e6BCA83a475C9f67f3484F7b3Cc3",
			ValueType: "string",
			Comment:   "Asset_Governance contract on Ethereum",
		},
		{
			Name:      Asset_Governance_Contract_Polygon,
			Value:     "0xCc1964a3cb6F26D8CF11c66fC7E3b9d3cbF07603",
			ValueType: "string",
			Comment:   "Asset_Governance contract on Polygon",
		},
		{
			Name:      Asset_Governance_Contract_NEAR_Aurora,
			Value:     "0xbd614B37042a8626581e38059CBcDcfDa0960E2C",
			ValueType: "string",
			Comment:   "Asset_Governance contract on Aurora",
		},
		{
			Name:      Asset_Governance_Contract_Avalanche,
			Value:     "0xb3ac75D65317b29368056ABecD1151808cCAd5B3",
			ValueType: "string",
			Comment:   "Asset_Governance contract on Avalanche",
		},
		{
			Name:      Asset_Governance_Contract_BSC,
			Value:     "0x3d92282d3678B3f5332d45646F4BB0a05Bdd83f6",
			ValueType: "string",
			Comment:   "Asset_Governance contract on BSC",
		},

		// Verifier Contract
		{
			Name:      Verifier_Contract_Ethereum,
			Value:     "0x2cC20EC93C126fbe647B010d49c01d1c8403be46",
			ValueType: "string",
			Comment:   "Verifier contract on Ethereum",
		},
		{
			Name:      Verifier_Contract_Polygon,
			Value:     "0x5704F78AA6E8a101F648F545a219984183696cd1",
			ValueType: "string",
			Comment:   "Verifier contract on Polygon",
		},
		{
			Name:      Verifier_Contract_NEAR_Aurora,
			Value:     "0xFbc7c80e779DF3DCef583BD106A2F5faCCa3c4Ed",
			ValueType: "string",
			Comment:   "Verifier contract on Aurora",
		},
		{
			Name:      Verifier_Contract_Avalanche,
			Value:     "0x5a9Bac44871cc225FC235b258EC26b0089Ab147f",
			ValueType: "string",
			Comment:   "Verifier contract on Avalanche",
		},
		{
			Name:      Verifier_Contract_BSC,
			Value:     "0xbfB7fE6Cf077c34c45FccBcb4666A373B974dd8F",
			ValueType: "string",
			Comment:   "Verifier contract on BSC",
		},
		// network rpc
		{
			Name:      "Ethereum_Test_Network_RPC",
			Value:     Ethereum_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Ethereum network rpc",
		},
		{
			Name:      "Polygon_Test_Network_RPC",
			Value:     Polygon_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Polygon network rpc",
		},
		{
			Name:      "Aurora_Test_Network_RPC",
			Value:     Aurora_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Aurora network rpc",
		},
		{
			Name:      "Avalanche_Test_Network_RPC",
			Value:     Avalanche_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Avalanche network rpc",
		},
		{
			Name:      "BSC_Test_Network_RPC",
			Value:     BSC_Test_Network_RPC,
			ValueType: "string",
			Comment:   "BSC network rpc",
		},
	}
}
