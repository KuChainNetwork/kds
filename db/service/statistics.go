package service

import (
	"fmt"

	"gorm.io/gorm"
	"kds/db/model"
)

// Statistics 统计
type Statistics struct {
	tableName string
}

// NewStatistics 工厂方法
func NewStatistics() *Statistics {
	return &Statistics{
		tableName: (&model.Statistics{}).TableName(),
	}
}

// Initialize 初始化
func (object *Statistics) Initialize(db *gorm.DB) (err error) {
	err = db.FirstOrCreate(&model.Statistics{}).Error
	return
}

// Load 加载
func (object *Statistics) Load(db *gorm.DB) (statistics *model.Statistics, err error) {
	var _statistics model.Statistics
	if err = db.Where("id=?", 1).
		Find(&_statistics).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	statistics = &_statistics
	return
}

// Updates 更新
func (object *Statistics) Updates(db *gorm.DB, updates map[string]interface{}) (err error) {
	err = db.Model(&model.Statistics{}).
		Where("id=?", 1).
		Updates(updates).Error
	return
}

// Increment 增加
func (object *Statistics) Increment(db *gorm.DB, key string, value int) (err error) {
	err = db.Model(&model.Statistics{}).
		Where("id=?", 1).
		Update(key, gorm.Expr(fmt.Sprintf(`%s + ?`, key), value)).Error
	return
}
