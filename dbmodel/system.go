package dbmodel

import (
	"encoding/json"

	"gorm.io/gorm"
)

// System 系统
type System struct {
	gorm.Model
	LastBlockHeight  int64 `json:"last_block_height" gorm:"last_block_height;type:bigint(20);not null"`
	IndexedBlockID   int64 `json:"indexed_block_id" gorm:"indexed_block_id;type:bigint(20);not null"`
	IndexedTxID      int64 `json:"indexed_tx_id" gorm:"indexed_tx_id;type:bigint(20);not null"`
	IndexedAddressID int64 `json:"indexed_address_id" gorm:"indexed_address_id;type:bigint(20);not null"`
	IndexedCoinID    int64 `json:"indexed_coin_id" gorm:"indexed_coin_id;type:bigint(20);not null"`
}

// TableName 表名
func (object *System) TableName() string {
	return "tb_system"
}

// String 字符串描述
func (object *System) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
