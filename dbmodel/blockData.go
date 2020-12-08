package dbmodel

import (
	"encoding/json"

	"gorm.io/gorm"
)

// BlockData 区块数据
type BlockData struct {
	gorm.Model
	Height  int64  `json:"height" gorm:"height;type:bigint(20);index:idx_height,unique,sort:desc;not null"`
	Block   []byte `json:"block" gorm:"block;type:json;not null"`
	Results []byte `json:"results" gorm:"results;type:json;not null"`
}

// TableName 表名
func (object *BlockData) TableName() string {
	return "tb_block_data"
}

// String 字符串描述
func (object *BlockData) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
