package dbservice

import (
	"kds/dbmodel"

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
func (object *BlockData) AddAll(db *gorm.DB, blocks []*dbmodel.BlockData) (err error) {
	err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(blocks).Error
	return
}

// List
func (object *BlockData) List(db *gorm.DB,
	startHeight,
	endHeight int64,
	mustContainsTx bool) (list []*dbmodel.BlockData, err error) {
	var arr []*dbmodel.BlockData
	if mustContainsTx {
		//db = db.Where("block->'$.block.data.txs' <> cast('null' as JSON)")
		db = db.Where("txn > 0")
	}
	if err = db.Model(&dbmodel.BlockData{}).Select("Height,Block,Results").
		Where("Height between ? and ?", startHeight, endHeight).
		Find(&arr).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	list = arr
	return
}
