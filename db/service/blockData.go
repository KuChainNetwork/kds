package service

import (
	"kds/db/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlockData
type BlockData struct {
}

// NewBlockData
func NewBlockData() *BlockData {
	return &BlockData{}
}

// AddAll
func (object *BlockData) AddAll(db *gorm.DB, blocks []*model.BlockData) (err error) {
	err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(blocks).Error
	return
}

// List
func (object *BlockData) List(db *gorm.DB,
	startHeight, limit int64, mustContainsTx bool) (list []*model.BlockData, err error) {
	var arr []*model.BlockData
	if mustContainsTx {
		db = db.Where("block->'$.block.data.txs' <> cast('null' as JSON)")
	}
	if err = db.Model(&model.BlockData{}).Select("Height,Block,Results").
		Order("Height ASC").
		Where("Height >= ?", startHeight).
		Limit(int(limit)).Find(&arr).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	list = arr
	return
}
