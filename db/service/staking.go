package service

import (
	"gorm.io/gorm"

	"kds/db/model"
)

// Staking
type Staking struct {
}

// NewStaking
func NewStaking() *Staking {
	return &Staking{}
}

// Add
func (object *Staking) Add(db *gorm.DB, staking *model.Staking) (err error) {
	err = db.Create(staking).Error
	return
}

// AddAll
func (object *Staking) AddAll(db *gorm.DB, list []*model.Staking) (err error) {
	err = db.Create(list).Error
	return
}
