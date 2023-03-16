package permctrl

import (
	"context"
	"errors"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/fetcher/address"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

var permissionControlConfig *PermissionControlConfig

type PermissionControl struct {
	fetcher *address.Fetcher
}

func NewPermissionControl(ctx context.Context, svcCtx *svc.ServiceContext) *PermissionControl {
	fetcher := address.NewFetcher(ctx, svcCtx)
	return &PermissionControl{
		fetcher: fetcher,
	}
}

func (c *PermissionControl) Control(txType uint32, txInfo string) error {
	l1Address, err := c.fetcher.GetL1AddressByTx(txType, txInfo)
	if err != nil {
		logx.Errorf("Can not get l1 address, txType:%d, txInfo:%s", txType, txInfo)
		return err
	}
	// If the permission control configuration has not been set
	// do not do the permission control config at all
	if permissionControlConfig == nil {
		return nil
	}
	// If the permission control switch is turned off
	// do not do the permission control config at all either
	if !permissionControlConfig.SwitchPermissionControlConfig {
		return nil
	}

	permissionControlItem := permissionControlConfig.GetPermissionControlConfigItem(txType)
	if permissionControlItem.PermissionControlType == ControlByWhitelist {
		if ok := containElement(l1Address, permissionControlItem.ControlWhiteList); !ok {
			return errors.New("the l1Address is not in the white list, can not do the transaction")
		}
	} else if permissionControlItem.PermissionControlType == ControlByBlacklist {
		if ok := containElement(l1Address, permissionControlItem.ControlBlackList); ok {
			return errors.New("the l1Address is in the black list, can not do the transaction")
		}
	}
	return nil
}

func containElement(element string, array []string) bool {
	for _, value := range array {
		if value == element {
			return true
		}
	}
	return false
}

func InitPermissionControl(config config.Config) {

	// Get the permission control configuration from the Apollo server
	permissionControlConfig = LoadApolloPermissionControlConfig(config)
	logx.Infof("Initiate Permission Control Facility Successfully!")
}
