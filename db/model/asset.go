package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Asset 资产
type Asset struct {
	gorm.Model
	Height int64     `json:"height" gorm:"height;type:bigint(20);not null"`
	TXHash string    `json:"tx_hash" gorm:"tx_hash;type:char(64);not null;index:idx_tx_hash,type:hash"`
	RealId string    `json:"real_id" gorm:"real_id;type:varchar(128);not null;index:idx_real_id"`
	Amount string    `json:"amount" gorm:"amount;type:varchar(128);not null"`
	Denom  string    `json:"denom" gorm:"denom;type:varchar(128);not null;index:idx_denom"`
	Time   time.Time `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Asset) TableName() string {
	return "tb_asset"
}

// String 字符串描述
func (object *Asset) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
