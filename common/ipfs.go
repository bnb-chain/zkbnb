package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mr-tron/base58/base58"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

type IPFS struct {
	shell *shell.Shell
}

var Ipfs *IPFS

func NewIPFS(url string) *IPFS {
	Ipfs = &IPFS{
		shell: shell.NewShell(url),
	}
	return Ipfs
}

func (i *IPFS) UploadOnlyHash(value string) (string, error) {
	cid, err := i.shell.Add(bytes.NewBufferString(value), shell.OnlyHash(true))
	if err != nil {
		return "", err
	}
	return cid, err
}

func (i *IPFS) Upload(value string) (string, error) {
	cid, err := i.shell.Add(bytes.NewBufferString(value))
	if err != nil {
		return "", err
	}
	return cid, err
}

func (i *IPFS) GenerateIPNS(ipnsName string) (*shell.Key, error) {
	return i.shell.KeyGen(context.Background(), ipnsName, shell.KeyGen.Type("ed25519"))
}

func (i *IPFS) PublishWithDetails(cid string, name string) (string, error) {
	cidPath := fmt.Sprintf("/%s/%s", "ipfs", cid)
	resp, err := i.shell.PublishWithDetails(cidPath, name, 0, 0, false)
	if err != nil {
		return "", err
	}
	if resp.Value != cidPath {
		logx.Severe(fmt.Sprintf("Expected to receive %s but got %s", cidPath, resp.Value))
		return "", errors.New(fmt.Sprintf("Expected to receive %s but got %s", cidPath, resp.Value))
	}
	return resp.Value, nil
}

func (i *IPFS) GenerateHash(cid string) (string, error) {
	base, err := base58.Decode(cid)
	if err != nil {
		return "", err
	}
	hex := hexutil.Encode(base)
	lowerHex := strings.ToLower(hex)
	return strings.Replace(lowerHex, "0x1220", "", 1), nil
}
