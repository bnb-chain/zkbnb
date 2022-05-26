package account

import (
	"gorm.io/gorm"
)

type AccountHistoryInfo struct {
	gorm.Model
	AccountIndex int64  `gorm:"index"`
	AccountName  string `gorm:"index"`
	PublicKey    string `gorm:"index"`
	BlockHeight  int64
	Status       int64
}
