package model

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Validator 验证人
type Validator struct {
	gorm.Model
	Height         int64           `json:"height" gorm:"height;type:bigint(20);not null"`
	Validator      string          `json:"validator" gorm:"validator;type:varchar(128);not null;index:idx_validator,unique"`
	Status         int             `json:"status" gorm:"status;type:tinyint(2);null;default 0"`
	Jailed         int             `json:"jailed" gorm:"jailed;type:tinyint;null;default 0"`
	Delegated      decimal.Decimal `json:"delegated" gorm:"delegated;type:decimal(38,18);null;default 0;index:idx_delegated_desc,sort:desc"`
	CommissionRate uint64          `json:"commission_rate" gorm:"commission_rate;type:bigint(20);not null"`
	Time           time.Time       `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Validator) TableName() string {
	return "tb_validator"
}

// String 字符串描述
func (object *Validator) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
