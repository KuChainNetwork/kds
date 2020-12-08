package dbservice

import (
	"gorm.io/gorm"

	"kds/dbmodel"
)

// Staking
type Staking struct {
}

// NewStaking
func NewStaking() *Staking {
	return &Staking{}
}

// Add
func (object *Staking) Add(db *gorm.DB, staking *dbmodel.Staking) (err error) {
	err = db.Create(staking).Error
	return
}

// AddAll
func (object *Staking) AddAll(db *gorm.DB, list []*dbmodel.Staking) (err error) {
	err = db.Create(list).Error
	return
}
