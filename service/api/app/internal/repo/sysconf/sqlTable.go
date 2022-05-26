package sysconf

import (
	"gorm.io/gorm"
)

type SysconfInfo struct {
	gorm.Model
	Name      string
	Value     string
	ValueType string
	Comment   string
}
