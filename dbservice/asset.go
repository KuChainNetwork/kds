package dbservice

import (
	"gorm.io/gorm"

	"kds/dbmodel"
)

type Asset struct{}

func NewAsset() *Asset {
	return &Asset{}
}
func (object *Asset) Add(db *gorm.DB, asset *dbmodel.Asset) (err error) {
	err = db.Create(asset).Error
	return
}
func (object *Asset) AddAll(db *gorm.DB, list []*dbmodel.Asset) (err error) {
	err = db.Create(list).Error
	return
}
