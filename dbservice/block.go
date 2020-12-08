package dbservice

import (
	"fmt"

	"gorm.io/gorm"

	"kds/dbmodel"
)

// Block
type Block struct {
	tableName string
}

// NewBlock
func NewBlock() *Block {
	return &Block{
		tableName: (&dbmodel.Block{}).TableName(),
	}
}

// LatestHeight
func (object *Block) LatestHeight(db *gorm.DB) (height int, err error) {
	sql := fmt.Sprintf(`select height from %s order by height desc limit 1`, object.tableName)
	if err = db.Raw(sql, &height).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// Latest
func (object *Block) Latest(db *gorm.DB) (m *dbmodel.Block, err error) {
	var list []*dbmodel.Block
	if err = db.Model(&dbmodel.Block{}).
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
func (object *Block) List(db *gorm.DB, offset, limit, sortByHeight int) (total int, list []*dbmodel.Block, err error) {
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
	var _list []*dbmodel.Block
	if 0 < sortByHeight {
		db = db.Order("Height ASC")
	} else if 0 > sortByHeight {
		db = db.Order("Height DESC")
	}
	if err = db.Offset(offset).
		Limit(limit).
		Select("Height", "Hash", "Txn", "Validator", "Time").
		Find(&_list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	list = _list
	return
}

// ListHeight
func (object *Block) ListHeight(db *gorm.DB, offset, limit int) (list []int64, err error) {
	if err = db.Model(&dbmodel.Block{}).
		Offset(offset).
		Limit(limit).
		Select("Height").
		Find(&list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// AddAll
func (object *Block) AddAll(db *gorm.DB, blocks []*dbmodel.Block) (err error) {
	err = db.Create(blocks).Error
	return
}
