package dbservice

import (
	"errors"

	"gorm.io/gorm"

	"kds/dbmodel"
)

// System system dbmodel dbservice
type System struct{}

// NewSystem factory method
func NewSystem() *System {
	return &System{}
}

// Initialize initialize system dbmodel
func (object *System) Initialize(db *gorm.DB) (m *dbmodel.System, err error) {
	s := &dbmodel.System{LastBlockHeight: 0}
	if err = db.FirstOrCreate(s).Error; nil == err || gorm.ErrRecordNotFound == err {
		m = s
		if nil != err {
			err = nil
		}
	}
	return
}

// UpdateLastBlockHeight
func (object *System) UpdateLastBlockHeight(db *gorm.DB,
	height int64) (err error) {
	return object.Updates(db, map[string]interface{}{
		"LastBlockHeight": height,
	})
}

// Updates 更新
func (object *System) Updates(db *gorm.DB, updates map[string]interface{}) (err error) {
	err = db.Model(&dbmodel.System{}).
		Where("id=?", 1).
		Updates(updates).Error
	return
}

func (object *System) GetLastBlockHeight(db *gorm.DB) (height int64) {
	err := db.Model(&dbmodel.System{}).
		Where("id=?", 1).Select("last_block_height").First(&height).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		height = 1
	}
	return
}
