package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Transfer 交易
type Transfer struct {
	gorm.Model
	TxHeight int64     `json:"tx_height" gorm:"tx_height;type:bigint(20);not null;index:idx_tx_height"`
	TxHash   string    `json:"tx_hash" gorm:"tx_hash;type:char(64);not null;index:idx_tx_hash"`
	Hash     string    `json:"hash" auth:"hash;type:char(64);not null;unique"`
	Auth     string    `json:"auth" gorm:"auth;type:varchar(128);not null"`
	From     string    `json:"from" gorm:"from;type:varchar(128);not null;index:idx_from;index:idx_from_to"`
	To       string    `json:"to" gorm:"to;type:varchar(128);not null;index:idx_to;index:idx_from_to"`
	Amount   string    `json:"amount" gorm:"amount;type:varchar(128);not null;index:idx_amount_denom"`
	Denom    string    `json:"denom" gorm:"denom;type:varchar(128);not null;index:idx_amount_denom"`
	Time     time.Time `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Transfer) TableName() string {
	return "tb_transfer"
}

// String 字符串描述
func (object *Transfer) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
