package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Staking 抵押
type Staking struct {
	gorm.Model
	Height        int64     `json:"height" gorm:"height;type:bigint(20);not null"`
	TXHash        string    `json:"tx_hash" gorm:"tx_hash;type:char(64);not null;index:idx_tx_hash,type:hash"`
	Validator     string    `json:"validator" gorm:"validator;type:varchar(128);not null;index:idx_validator;index:idx_validator_delegator"`
	Delegator     string    `json:"delegator" gorm:"delegator;type:varchar(128);not null;index:idx_delegator;index:idx_validator_delegator"`
	StakingAmount string    `json:"staking_amount" gorm:"staking_amount;type:varchar(128);not null"`
	StakingDenom  string    `json:"staking_denom" gorm:"staking_denom;type:varchar(128);not null"`
	Time          time.Time `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Staking) TableName() string {
	return "tb_staking"
}

// String 字符串描述
func (object *Staking) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
