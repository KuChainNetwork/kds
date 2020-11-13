package service

import (
	"gorm.io/gorm"

	"kds/db/model"
)

type Asset struct{}

func NewAsset() *Asset {
	return &Asset{}
}
func (object *Asset) Add(db *gorm.DB, asset *model.Asset) (err error) {
	err = db.Create(asset).Error
	return
}
func (object *Asset) AddAll(db *gorm.DB, list []*model.Asset) (err error) {
	err = db.Create(list).Error
	return
}
