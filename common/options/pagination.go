package options

import "gorm.io/gorm"

func Offset(offset int) QueryOption {
	return func(db *gorm.DB) {
		db.Offset(offset)
	}
}

func Limit(limit int) QueryOption {
	return func(db *gorm.DB) {
		db.Limit(limit)
	}
}
