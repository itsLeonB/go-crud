package internal

import "gorm.io/gorm"

type DeletedFilterType interface {
	WhereDeleted() func(*gorm.DB) *gorm.DB
}

type ExcludeDeleted struct{}

func (ed ExcludeDeleted) WhereDeleted() func(*gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		return d.Where("deleted_at IS NULL")
	}
}

type IncludeDeleted struct{}

func (id IncludeDeleted) WhereDeleted() func(*gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		return d
	}
}

type OnlyDeleted struct{}

func (od OnlyDeleted) WhereDeleted() func(*gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		return d.Where("deleted_at IS NOT NULL")
	}
}
