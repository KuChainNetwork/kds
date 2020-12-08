package dbservice

import (
	"fmt"

	"gorm.io/gorm"

	"kds/dbmodel"
)

// TX
type TX struct {
	tableName string
}

// NewTX
func NewTX() *TX {
	return &TX{
		tableName: (&dbmodel.TX{}).TableName(),
	}
}

// Latest
func (object *TX) Latest(db *gorm.DB) (m *dbmodel.TX, err error) {
	var list []*dbmodel.TX
	if err = db.Model(&dbmodel.TX{}).
		Order("Height DESC").
		Limit(1).
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
func (object *TX) List(db *gorm.DB, offset, limit, sortByHeight int) (total int, list []*dbmodel.TX, err error) {
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
	var _list []*dbmodel.TX
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
	if err = db.Model(&dbmodel.TX{}).
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
func (object *TX) Add(db *gorm.DB, tx *dbmodel.TX) (err error) {
	err = db.Create(tx).Error
	return
}

// AddAll
func (object *TX) AddAll(db *gorm.DB, list []*dbmodel.TX) (err error) {
	err = db.Create(list).Error
	return
}
