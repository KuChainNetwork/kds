package service

import (
	"fmt"

	"gorm.io/gorm"

	"kds/db/model"
)

// TX
type TX struct {
	tableName string
}

// NewTX
func NewTX() *TX {
	return &TX{
		tableName: (&model.TX{}).TableName(),
	}
}

// Latest
func (object *TX) Latest(db *gorm.DB) (m *model.TX, err error) {
	var list []*model.TX
	if err = db.Model(&model.TX{}).
		Order("Height DESC").
		Find(&list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	if 0 < len(list) {
		m = list[0]
	}
	return
}

// List
func (object *TX) List(db *gorm.DB, offset, limit, sortByHeight int) (total int, list []*model.TX, err error) {
	var count int
	if err = db.Raw(fmt.Sprintf(`select count(id) as count from %s`, object.tableName)).Find(&count).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	if 0 >= count {
		return
	}
	total = count
	var _list []*model.TX
	if 0 < sortByHeight {
		db = db.Order("Height ASC")
	} else if 0 > sortByHeight {
		db = db.Order("Height DESC")
	}
	if err = db.Offset(offset).
		Limit(limit).
		Select("Hash", "Height", "Time", "Status", "From", "RealTo", "Amount", "Denom").
		Find(&_list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	list = _list
	return
}

// ListHash
func (object *TX) ListHash(db *gorm.DB, offset, limit int) (list []string, err error) {
	list = nil
	if err = db.Model(&model.TX{}).
		Offset(offset).
		Limit(limit).
		Select("Hash").
		Find(&list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// Add
func (object *TX) Add(db *gorm.DB, tx *model.TX) (err error) {
	err = db.Create(tx).Error
	return
}

// AddAll
func (object *TX) AddAll(db *gorm.DB, list []*model.TX) (err error) {
	err = db.Create(list).Error
	return
}
