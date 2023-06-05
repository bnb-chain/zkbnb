package permctrl

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/fetcher/address"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionControl struct {
	fetcher *address.Fetcher
}

func NewPermissionControl(svcCtx *svc.ServiceContext) *PermissionControl {
	fetcher := address.NewFetcher(svcCtx)
	return &PermissionControl{
		fetcher: fetcher,
	}
}

func (c *PermissionControl) Control(txType uint32, txInfo string) error {
	l1AddressArray, err := c.fetcher.GetL1AddressByTx(txType, txInfo)
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
		if ok := containElements(l1AddressArray, permissionControlItem.ControlWhiteList); !ok {
			return types.AppErrPermissionForbidden
		}
	} else if permissionControlItem.PermissionControlType == ControlByBlacklist {
		if ok := containElements(l1AddressArray, permissionControlItem.ControlBlackList); ok {
			return types.AppErrPermissionForbidden
		}
	}
	return nil
}

func containElements(elementArray []string, array []string) bool {
	for _, element := range elementArray {
		for _, value := range array {
			if value == element {
				return true
			}
		}
	}
	return false
}
