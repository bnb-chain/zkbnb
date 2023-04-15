package flags

import (
	"github.com/urfave/cli/v2"
)

var (
	ConfigFlag = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"f"},
		Usage:   "the config file",
	}
	ContractAddrFlag = &cli.StringFlag{
		Name:  "contractAddr",
		Usage: "the contract addresses file",
	}
	DSNFlag = &cli.StringFlag{
		Name:  "dsn",
		Usage: "data source name",
	}
	BSCTestNetworkRPCFlag = &cli.StringFlag{
		Name:  "testnet",
		Value: "https://bsc-testnet.nodereal.io/v1/a1cee760ac744f449416a711f20d99dd",
		Usage: "the rpc endpoint of bsc testnet",
	}
	LocalTestNetworkRPCFlag = &cli.StringFlag{
		Name:  "local",
		Value: "http://127.0.0.1:8545/",
		Usage: "the rpc endpoint of local net",
	}
	BlockHeightFlag = &cli.Int64Flag{
		Name:  "height",
		Usage: "block height",
	}
	ServiceNameFlag = &cli.StringFlag{
		Name:  "service",
		Usage: "service name(committer, witness)",
	}
	BatchSizeFlag = &cli.IntFlag{
		Name:  "batch",
		Value: 1000,
		Usage: "batch size for reading history record from the database",
	}
	PProfEnabledFlag = &cli.BoolFlag{
		Name:  "pprof",
		Value: false,
		Usage: "Enable the pprof HTTP server",
	}
	PProfPortFlag = &cli.IntFlag{
		Name:  "pprof.port",
		Usage: "pprof HTTP server listening port",
		Value: 6060,
	}
	PProfAddrFlag = &cli.StringFlag{
		Name:  "pprof.addr",
		Usage: "pprof HTTP server listening interface",
		Value: "127.0.0.1",
	}
	MetricsEnabledFlag = &cli.BoolFlag{
		Name:  "metrics",
		Value: false,
		Usage: "Enable metrics collection and reporting",
	}
	MetricsHTTPFlag = &cli.StringFlag{
		Name:  "metrics.addr",
		Usage: "Enable stand-alone metrics HTTP server listening interface",
		Value: "127.0.0.1",
	}
	MetricsPortFlag = &cli.IntFlag{
		Name:  "metrics.port",
		Usage: "Metrics HTTP server listening port",
		Value: 6060,
	}
	CommandFlag = &cli.StringFlag{
		Name:    "command",
		Aliases: []string{"m"},
		Usage:   "command",
	}
	RecoveryFromHistoryFlag = &cli.BoolFlag{
		Name:  "recoveryFromHistory",
		Value: true,
		Usage: "read account and nft history record from the database, or read account and nft from the database",
	}
	PrivateKeyFlag = &cli.StringFlag{
		Name:  "privateKey",
		Usage: "private key",
		Value: "",
	}
	TokenFlag = &cli.StringFlag{
		Name:  "token",
		Usage: "token",
		Value: "",
	}
	AmountFlag = &cli.StringFlag{
		Name:  "amount",
		Usage: "amount",
		Value: "",
	}
	ProofFlag = &cli.StringFlag{
		Name:  "proof",
		Usage: "proof",
		Value: "",
	}
	NftIndexListFlag = &cli.StringFlag{
		Name:  "nftIndexList",
		Usage: "nftIndexList",
		Value: "",
	}
	AddressFlag = &cli.StringFlag{
		Name:  "address",
		Usage: "address",
		Value: "",
	}
	ProofFolderFlag = &cli.StringFlag{
		Name:  "proofFolder",
		Usage: "proofFolder",
		Value: "",
	}
	RevertBlockHeightFlag = &cli.Int64Flag{
		Name:  "height",
		Usage: "height",
		Value: 0,
	}
	RollbackBlockHeightFlag = &cli.Int64Flag{
		Name:  "height",
		Usage: "height",
		Value: 0,
	}
)
