package dbservice

import (
	"fmt"

	"gorm.io/gorm"

	"kds/dbmodel"
)

// Coin 代币
type Coin struct {
	tableName string
}

// NewCoin 工厂方法
func NewCoin() *Coin {
	return &Coin{
		tableName: (&dbmodel.Coin{}).TableName(),
	}
}

// List 列表
func (object *Coin) List(db *gorm.DB,
	offset, limit int) (total int, list []*dbmodel.Coin, err error) {
	var count int
	if err = db.Raw(fmt.Sprintf(`select count(id) as count from %s`, object.tableName)).Find(&count).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	if 0 >= count {
		return
	}
	total = count
	var _list []*dbmodel.Coin
	if err = db.Offset(offset).
		Limit(limit).
		Select("Symbol",
			"Description",
			"Time",
			"MaxSupplyAmount",
			"MaxSupplyDenom",
			"IssueAmount",
			"IssueDenom",
			"Creator").
		Find(&_list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	list = _list
	return
}

// ListSymbol 列表符号
func (object *Coin) ListSymbol(db *gorm.DB, offset, limit int) (list []string, err error) {
	if err = db.Model(&dbmodel.Coin{}).
		Offset(offset).
		Limit(limit).
		Select("Symbol").
		Find(&list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// LikeSymbol
func (object *Coin) LikeSymbol(db *gorm.DB, key string, offset, limit int) (list []string, err error) {
	if err = db.Model(&dbmodel.Coin{}).
		Where(fmt.Sprintf(`symbol like '%s%%'`, key)).
		Offset(offset).
		Limit(limit).
		Select("symbol").
		Distinct("symbol").
		Find(&list).Error; nil != err {
		if gorm.ErrRecordNotFound == err {
			err = nil
		}
		return
	}
	return
}

// Add 添加
func (object *Coin) Add(db *gorm.DB, coin *dbmodel.Coin) (err error) {
	err = db.Create(coin).Error
	return
}

// AddAll 添加所有
func (object *Coin) AddAll(db *gorm.DB, list []*dbmodel.Coin) (err error) {
	err = db.Create(list).Error
	return
}

// UpdateIssue 更新流通量
func (object *Coin) UpdateIssue(db *gorm.DB, coin *dbmodel.Coin) (err error) {
	err = db.Model(&dbmodel.Coin{}).
		Where("Creator=?", coin.Creator).
		Where("Symbol=?", coin.Symbol).
		Updates(map[string]interface{}{
			"IssueAmount": coin.IssueAmount,
			"IssueDenom":  coin.IssueDenom,
		}).Error
	return
}
