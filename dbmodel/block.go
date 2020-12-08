package dbmodel

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Block 区块
type Block struct {
	gorm.Model
	Height    int64     `json:"height" gorm:"height;type:bigint(20);index:idx_height_unique,unique;index:idx_height_desc,sort:desc;not null"`
	Hash      string    `json:"hash" gorm:"hash;type:char(64);not null"`
	Txn       int64     `json:"txn" gorm:"txn;type:bigint(20);not null"`
	Validator string    `json:"validator" gorm:"validator;type:varchar(128);not null"`
	Time      time.Time `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Block) TableName() string {
	return "tb_block"
}

// String 字符串描述
func (object *Block) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
