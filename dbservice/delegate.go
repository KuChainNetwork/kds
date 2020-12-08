package dbservice

import (
	"fmt"

	"gorm.io/gorm"

	"kds/dbmodel"
)

// Delegate 抵押
type Delegate struct {
	tableName string
}

// NewDelegate 工厂方法
func NewDelegate() *Delegate {
	return &Delegate{
		tableName: (&dbmodel.Delegate{}).TableName(),
	}
}

// List 列表
func (object *Delegate) List(db *gorm.DB, offset, limit int) (total int, list []*dbmodel.Delegate, err error) {
	var count int
	sql := fmt.Sprintf(`select count(id) as count from %s group by validator`,
		object.tableName)
	if err = db.Raw(sql).Find(&count).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	if 0 >= count {
		return
	}
	total = count
	sql = fmt.Sprintf(`select tb_delegate.validator, tb_delegate.amount, tb_delegate.denom
from (select validator, sum(amount) as amount, denom
from %s
group by validator, denom
limit %d offset %d) tb_delegate`,
		object.tableName,
		limit,
		offset)
	var _list []*dbmodel.Delegate
	if err = db.Raw(sql).Find(&_list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	list = _list
	return
}

// Add 添加
func (object *Delegate) Add(db *gorm.DB, delegate *dbmodel.Delegate) (err error) {
	err = db.Create(delegate).Error
	return
}

// AddAll 添加所有
func (object *Delegate) AddAll(db *gorm.DB, list []*dbmodel.Delegate) (err error) {
	err = db.Create(list).Error
	return
}
