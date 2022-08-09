package options

import "gorm.io/gorm"

type QueryOption func(*gorm.DB)
