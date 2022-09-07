package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/zkbnb/cmd/flags"
	"github.com/bnb-chain/zkbnb/service/apiserver"
	"github.com/bnb-chain/zkbnb/service/committer"
	"github.com/bnb-chain/zkbnb/service/monitor"
	"github.com/bnb-chain/zkbnb/service/prover"
	"github.com/bnb-chain/zkbnb/service/sender"
	"github.com/bnb-chain/zkbnb/service/witness"
	"github.com/bnb-chain/zkbnb/tools/dbinitializer"
	"github.com/bnb-chain/zkbnb/tools/recovery"
)

// Build Info (set via linker flags)
var (
	gitCommit = "unknown"
	gitDate   = "unknown"
	version   = "unknown"
)

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		fmt.Println("Version:", ctx.App.Version)
		fmt.Println("Git Commit:", gitCommit)
		fmt.Println("Git Commit Date:", gitDate)
		fmt.Println("Architecture:", runtime.GOARCH)
		fmt.Println("Go Version:", runtime.Version())
		fmt.Println("Operating System:", runtime.GOOS)
	}

	app := &cli.App{
		Name:        "ZkBNB",
		HelpName:    "zkbnb",
		Version:     version,
		Description: "ZkRollup BNB Application Side Chain",
		Commands: []*cli.Command{
			// services
			{
				Name:  "prover",
				Usage: "Run prover service",
				Flags: []cli.Flag{
					flags.ConfigFlag,
				},
				Action: func(cCtx *cli.Context) error {
					if !cCtx.IsSet(flags.ConfigFlag.Name) {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return prover.Run(cCtx.String(flags.ConfigFlag.Name))
				},
			},
			{
				Name:  "witness",
				Usage: "Run witness service",
				Flags: []cli.Flag{
					flags.ConfigFlag,
				},
				Action: func(cCtx *cli.Context) error {
					if !cCtx.IsSet(flags.ConfigFlag.Name) {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return witness.Run(cCtx.String(flags.ConfigFlag.Name))
				},
			},
			{
				Name:  "monitor",
				Usage: "Run monitor service",
				Flags: []cli.Flag{
					flags.ConfigFlag,
				},
				Action: func(cCtx *cli.Context) error {
					if !cCtx.IsSet(flags.ConfigFlag.Name) {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return monitor.Run(cCtx.String(flags.ConfigFlag.Name))
				},
			},
			{
				Name: "committer",
				Flags: []cli.Flag{
					flags.ConfigFlag,
				},
				Usage: "Run committer service",
				Action: func(cCtx *cli.Context) error {
					if !cCtx.IsSet(flags.ConfigFlag.Name) {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return committer.Run(cCtx.String(flags.ConfigFlag.Name))
				},
			},
			{
				Name:  "sender",
				Usage: "Run sender service",
				Flags: []cli.Flag{
					flags.ConfigFlag,
				},
				Action: func(cCtx *cli.Context) error {
					if !cCtx.IsSet(flags.ConfigFlag.Name) {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return sender.Run(cCtx.String(flags.ConfigFlag.Name))
				},
			},
			{
				Name:  "apiserver",
				Usage: "Run apiserver service",
				Flags: []cli.Flag{
					flags.ConfigFlag,
				},
				Action: func(cCtx *cli.Context) error {
					if !cCtx.IsSet(flags.ConfigFlag.Name) {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return apiserver.Run(cCtx.String(flags.ConfigFlag.Name))
				},
			},
			// tools
			{
				Name:  "db",
				Usage: "Database tools",
				Subcommands: []*cli.Command{
					{
						Name:  "initialize",
						Usage: "Initialize DB tables",
						Flags: []cli.Flag{
							flags.ContractAddrFlag,
							flags.DSNFlag,
							flags.BSCTestNetworkRPCFlag,
							flags.LocalTestNetworkRPCFlag,
						},
						Action: func(cCtx *cli.Context) error {
							if !cCtx.IsSet(flags.ContractAddrFlag.Name) ||
								!cCtx.IsSet(flags.DSNFlag.Name) {
								return cli.ShowSubcommandHelp(cCtx)
							}

							return dbinitializer.Initialize(
								cCtx.String(flags.DSNFlag.Name),
								cCtx.String(flags.ContractAddrFlag.Name),
								cCtx.String(flags.BSCTestNetworkRPCFlag.Name),
								cCtx.String(flags.LocalTestNetworkRPCFlag.Name),
							)
						},
					},
				},
			},
			{
				Name:  "tree",
				Usage: "TreeDB tools",
				Subcommands: []*cli.Command{
					{
						Name:  "recovery",
						Usage: "Recovery treedb from the database",
						Flags: []cli.Flag{
							flags.ConfigFlag,
							flags.BlockHeightFlag,
							flags.ServiceNameFlag,
							flags.BatchSizeFlag,
						},
						Action: func(cCtx *cli.Context) error {
							if !cCtx.IsSet(flags.ServiceNameFlag.Name) ||
								!cCtx.IsSet(flags.BlockHeightFlag.Name) ||
								!cCtx.IsSet(flags.ConfigFlag.Name) {
								return cli.ShowSubcommandHelp(cCtx)
							}
							recovery.RecoveryTreeDB(
								cCtx.String(flags.ConfigFlag.Name),
								cCtx.Int64(flags.BlockHeightFlag.Name),
								cCtx.String(flags.ServiceNameFlag.Name),
								cCtx.Int(flags.BatchSizeFlag.Name),
							)
							return nil
						},
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
