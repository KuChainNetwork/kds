package dbservice

import (
	"gorm.io/gorm"

	"kds/dbmodel"
)

// Transfer
type Transfer struct {
}

// NewTransfer
func NewTransfer() *Transfer {
	return &Transfer{}
}

// Add
func (object *Transfer) Add(db *gorm.DB, tx *dbmodel.Transfer) (err error) {
	err = db.Create(tx).Error
	return
}

// AddAll
func (object *Transfer) AddAll(db *gorm.DB, list []*dbmodel.Transfer) (err error) {
	err = db.Create(list).Error
	return
}
