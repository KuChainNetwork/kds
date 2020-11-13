package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Coin 币
type Coin struct {
	gorm.Model
	Height          int64     `json:"height" gorm:"height;type:bigint(20);not null"`
	TXHash          string    `json:"tx_hash" gorm:"tx_hash;type:char(64);not null;index:idx_tx_hash,type:hash"`
	Creator         string    `json:"creator" gorm:"creator;type:varchar(128);not null"`
	Symbol          string    `json:"symbol" gorm:"symbol;type:varchar(128);not null;uniqueKey:idx_name"`
	MaxSupplyAmount string    `json:"max_supply_amount" gorm:"max_supply_amount;type:varchar(128);not null"`
	MaxSupplyDenom  string    `json:"max_supply_denom" gorm:"max_supply_denom;type:varchar(128);not null"`
	IssueAmount     string    `json:"issue_amount" gorm:"issue_amount;type:varchar(128);null"`
	IssueDenom      string    `json:"issue_denom" gorm:"issue_denom;type:varchar(128);null"`
	Description     string    `json:"description" gorm:"description;type:varchar(256);null"`
	Time            time.Time `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Coin) TableName() string {
	return "tb_coin"
}

// String 字符串描述
func (object *Coin) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
