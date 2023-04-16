package permctrl

import (
	"github.com/bnb-chain/zkbnb/types"
)

type PermissionCheck struct {
}

func NewPermissionCheck() *PermissionCheck {
	return &PermissionCheck{}
}

func (p *PermissionCheck) CheckPermission(l1Address string, txType uint32) (bool, error) {

	// If the permission control configuration has not been set
	// return true by default, this means the account is permitted by default
	if permissionControlConfig == nil {
		return true, nil
	}
	// If the permission control switch is turned off
	// return true by default, this means the account is permitted by default
	if !permissionControlConfig.SwitchPermissionControlConfig {
		return true, nil
	}

	permissionControlItem := permissionControlConfig.GetPermissionControlConfigItem(txType)
	if permissionControlItem.PermissionControlType == ControlByWhitelist {
		if ok := containElement(l1Address, permissionControlItem.ControlWhiteList); !ok {
			return false, types.AppErrPermissionForbidden
		}
	} else if permissionControlItem.PermissionControlType == ControlByBlacklist {
		if ok := containElement(l1Address, permissionControlItem.ControlBlackList); ok {
			return false, types.AppErrPermissionForbidden
		}
	}
	return true, nil
}

func containElement(l1Address string, array []string) bool {
	for _, value := range array {
		if value == l1Address {
			return true
		}
	}
	return false
}
