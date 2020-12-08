package dbservice

import (
	"fmt"

	"gorm.io/gorm"

	"kds/dbmodel"
)

// Validator 验证人
type Validator struct {
	tableName string
}

// NewValidator 工厂方法
func NewValidator() *Validator {
	return &Validator{
		tableName: (&dbmodel.Validator{}).TableName(),
	}
}

// List 列表
func (object *Validator) List(db *gorm.DB, status, offset, limit int) (total int, list []*dbmodel.Validator, err error) {
	var count int
	sql := fmt.Sprintf(`select count(id) as count from %s where status = ?`, object.tableName)
	if err = db.Raw(sql, status).Find(&count).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	if 0 >= count {
		return
	}
	total = count
	var _list []*dbmodel.Validator
	if err = db.Select("Validator", "Delegated", "CommissionRate").
		Order("Delegated DESC").
		Offset(offset).
		Limit(limit).
		Find(&_list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	list = _list
	return
}

// Count 总数
func (object *Validator) Count(db *gorm.DB) (total int, err error) {
	sql := fmt.Sprintf(`select count(id) as count from %s`, object.tableName)
	if err = db.Raw(sql).Find(&total).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// Add 添加
func (object *Validator) Add(db *gorm.DB, validator *dbmodel.Validator) (err error) {
	err = db.Create(validator).Error
	return
}

// AddAll 添加所有
func (object *Validator) AddAll(db *gorm.DB, list []*dbmodel.Validator) (err error) {
	err = db.Create(list).Error
	return
}

// UpdateStatus 更新状态
func (object *Validator) UpdateStatus(db *gorm.DB, validator *dbmodel.Validator) (err error) {
	err = db.Model(&dbmodel.Validator{}).
		Where("validator=?", validator).
		Update("status", validator.Status).Error
	return
}

// UpdateJailed 释放
func (object *Validator) UpdateJailed(db *gorm.DB,
	validator string,
	jailed bool) (err error) {
	err = db.Model(&dbmodel.Validator{}).
		Where("validator=?", validator).
		Update("jailed", func() int {
			if jailed {
				return 1
			}
			return 0
		}()).Error
	return
}
