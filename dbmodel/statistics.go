package dbmodel

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Statistics 统计
type Statistics struct {
	gorm.Model
	TotalValidator int `json:"total_validator" gorm:"total_validator;type:bigint(20);not null;default 0"`
	LatestHeight   int `json:"latest_height" gorm:"latest_height;type:bigint(20);not null;default 0"`
	TotalTX        int `json:"total_tx" gorm:"total_tx;type:bigint(20);not null;default 0"`
	TotalAccount   int `json:"total_account" gorm:"total_account;type:bigint(20);not null;default 0"`
}

// TableName 表名
func (object *Statistics) TableName() string {
	return "tb_statistics"
}

// String 字符串描述
func (object *Statistics) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
