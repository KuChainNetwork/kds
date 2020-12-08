package dbmodel

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// TX 交易
type TX struct {
	gorm.Model
	Hash        string    `json:"hash" gorm:"hash;type:char(64);not null"`
	Height      int64     `json:"height" gorm:"height;type:bigint(20);not null;index:idx_height_desc,sort:desc"`
	Route       string    `json:"route" gorm:"route;type:varchar(128);index:idx_route,type:btree;not null"`
	Type        string    `json:"type" gorm:"type;type:varchar(64);index:idx_type,type:btree;index:idx_type_from,type:btree;not null"`
	From        string    `json:"from" gorm:"from;type:varchar(128);index:idx_from,type:btree;index:idx_type_from,type:btree;not null"`
	To          string    `json:"to" gorm:"to;type:varchar(128);index:idx_to,type:btree;not null"`
	RealTo      string    `json:"real_to" gorm:"real_to;type:varchar(128);null"`
	Time        time.Time `json:"time" gorm:"time;type:timestamp;not null"`
	Amount      string    `json:"amount" gorm:"amount;type:varchar(128);not null"`
	Denom       string    `json:"denom" gorm:"denom;type:varchar(128);not null"`
	Code        uint32    `json:"code" gorm:"code;type:smallint;not null;index:idx_code"`
	Log         string    `json:"log" gorm:"log;type:text;null"`
	Info        string    `json:"info" gorm:"info;type:text;null"`
	GasWanted   int64     `json:"gas_wanted" gorm:"gas_wanted;type:bigint(20);not null"`
	GasUsed     int64     `json:"gas_used" gorm:"gas_used;type:bigint(20);not null"`
	Events      []byte    `json:"events" gorm:"events;type:JSON;not null"`
	Message     []byte    `json:"message" gorm:"message;type:JSON;not null"`
	MessageData []byte    `json:"message_data" gorm:"message_data;type:JSON;not null"`
}

// TableName 表名
func (object *TX) TableName() string {
	return "tb_tx"
}

// String 字符串描述
func (object *TX) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
