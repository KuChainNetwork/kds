package db

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang/glog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kds/dbmodel"
	"kds/singleton"
)

var (
	initializeOnce sync.Once // 数据库初始化一次
)

// gLogWrapper
type gLogWrapper struct {
}

// Write 写原始消息
func (object *gLogWrapper) Write(p []byte) (n int, err error) {
	glog.Warning(string(p))
	n = len(p)
	return
}

// connect 连接数据库
func connect(dsn string, retryTimes int) (err error) {
	newLogger := logger.New(
		log.New(&gLogWrapper{}, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Warn, // Log level
			Colorful:      false,       // Disable color
		},
	)
	for i := 0; i < retryTimes; i++ {
		if singleton.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger}); nil == err {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return
}

// migrate 重构表格
func migrate() (err error) {
	err = singleton.DB.AutoMigrate(
		&dbmodel.Account{},
		&dbmodel.Asset{},
		&dbmodel.Block{},
		&dbmodel.BlockData{},
		&dbmodel.Coin{},
		&dbmodel.Delegate{},
		&dbmodel.Staking{},
		&dbmodel.Statistics{},
		&dbmodel.System{},
		&dbmodel.Transfer{},
		&dbmodel.TX{},
		&dbmodel.Validator{},
	)
	return
}

// Initialize 初始化数据库单例
func Initialize(username, password, host, database string,
	port, retryTimes int) (err error) {
	initializeOnce.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			username,
			password,
			host,
			port,
			database)
		for _, fn := range []func() error{
			func() error { return connect(dsn, retryTimes) },
			migrate,
		} {
			if err = fn(); nil != err {
				return
			}
		}
	})
	return
}

// Dispose 销毁数据库单例
func Dispose() {
	mysqlDB, err := singleton.DB.DB()
	if nil != err {
		glog.Fatalln(err)
		return
	}
	mysqlDB.Close()
}
