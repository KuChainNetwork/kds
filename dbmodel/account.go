package dbmodel

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Account 账户
type Account struct {
	gorm.Model
	Height    int64     `json:"height" gorm:"height;type:bigint(20);not null"`
	TXHash    string    `json:"tx_hash" gorm:"tx_hash;type:char(64);not null;index:id_tx_hash,type:hash"`
	Creator   string    `json:"creator" gorm:"creator;type:varchar(128);not null"`
	AccountID string    `json:"account_id" gorm:"account_id;type:varchar(128);not null"`
	Number    uint64    `json:"number" gorm:"number;type:bigint(20);not null"`
	Name      string    `json:"name" gorm:"name;type:varchar(128);not null;uniqueIndex:idx_name_auth;index:idx_name"`
	Auth      string    `json:"auth" gorm:"auth:type:varchar(128);not null;uniqueIndex:idx_name_auth;index:idx_auth"`
	Time      time.Time `json:"time" gorm:"time;type:timestamp;not null"`
}

// TableName 表名
func (object *Account) TableName() string {
	return "tb_account"
}

// String 字符串描述
func (object *Account) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
