package dbmodel

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Delegate 抵押
type Delegate struct {
	gorm.Model
	Height    int64           `json:"height" gorm:"height;type:bigint(20);not null"`
	TXHash    string          `json:"tx_hash" gorm:"tx_hash;type:char(64);not null;index:idx_tx_hash,type:hash"`
	Delegator string          `json:"delegator" gorm:"delegator;type:varchar(128);not null;index:idx_delegator;index:idx_delegator_validator"`
	Validator string          `json:"validator" gorm:"validator;type:varchar(128);not null;index:idx_validator;index:idx_delegator_validator;index:idx_validator_denom"`
	Amount    decimal.Decimal `json:"amount" gorm:"amount;type:decimal(38,18);not null"`
	Denom     string          `json:"denom" gorm:"denom;type:varchar(128);not null;index:idx_validator_denom"`
	Time      time.Time       `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Delegate) TableName() string {
	return "tb_delegate"
}

// String 字符串描述
func (object *Delegate) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
