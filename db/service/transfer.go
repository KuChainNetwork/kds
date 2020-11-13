package service

import (
	"gorm.io/gorm"

	"kds/db/model"
)

// Transfer
type Transfer struct {
}

// NewTransfer
func NewTransfer() *Transfer {
	return &Transfer{}
}

// Add
func (object *Transfer) Add(db *gorm.DB, tx *model.Transfer) (err error) {
	err = db.Create(tx).Error
	return
}

// AddAll
func (object *Transfer) AddAll(db *gorm.DB, list []*model.Transfer) (err error) {
	err = db.Create(list).Error
	return
}
