package service

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"kds/db/model"
)

// Account 账户
type Account struct {
}

// NewAccount 工厂方法
func NewAccount() *Account {
	return &Account{}
}

// ListAddress 列表地址
func (object *Account) ListAddress(db *gorm.DB, offset, limit int) (list []string, err error) {
	if err = db.Model(&model.Account{}).
		Offset(offset).
		Limit(limit).
		Select("Creator").
		Find(&list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// LikeAddress 搜索地址
func (object *Account) LikeAddress(db *gorm.DB, key string, offset, limit int) (list []string, err error) {
	if err = db.Model(&model.Account{}).
		Where(fmt.Sprintf(`creator like '%s%%'`, key)).
		Offset(offset).
		Limit(limit).
		Select("creator").
		Distinct("creator").
		Find(&list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// Add 添加
func (object *Account) Add(db *gorm.DB, account *model.Account) (err error) {
	err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(account).Error // 已存在不执行操作
	return
}

// AddAll 添加所有
func (object *Account) AddAll(db *gorm.DB, list []*model.Account) (err error) {
	err = db.Create(list).Error
	return
}

// UpdateAuth 更新公钥
func (object *Account) UpdateAuth(db *gorm.DB, auth *model.Account) (err error) {
	err = db.Model(&model.Account{}).
		Where("Name=?", auth.Name).
		UpdateColumn("Auth=?", auth.Auth).Error
	return
}
