package generateproof

import (
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/config"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/generateproof"
	"github.com/goccy/go-json"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

const CommandRunGenerateProof = "run"
const CommandContinueGenerateProof = "continue"

func Run(configFile string, address string, token string, nftIndexListStr string, proofFolder string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	if address != "" {
		c.Address = address
	}
	if token != "" {
		c.Token = token
	}
	if proofFolder != "" {
		c.ProofFolder = proofFolder
	}
	if nftIndexListStr != "" {
		var nftIndexList []int64
		err := json.Unmarshal([]byte(nftIndexListStr), &nftIndexList)
		if err != nil {
			return nil
		}
		c.NftIndexList = nftIndexList
	}

	m, err := generateproof.NewMonitor(&c)
	if err != nil {
		logx.Severe(err)
		panic(err)
	}

	go func() {
		for {
			err := m.MonitorGenericBlocks()
			if err != nil {
				logx.Severef("monitor blocks error, %v", err)
			} else {
				break
			}
		}
	}()

	exodusExit, err := generateproof.NewExodusExit(&c)
	if err != nil {
		return err
	}
	err = exodusExit.Run()
	if err != nil {
		return err
	}
	logx.Info("generateproof is starting......")
	return nil
}
