package db

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kds/db/model"
	"kds/db/service"
	"kds/singleton"
)

var (
	initializeOnce sync.Once // 数据库初始化一次
)

// connect 连接数据库
func connect(dsn string, retryTimes int) (err error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Warn, // Log level
			Colorful:      true,        // Disable color
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
		&model.Account{},
		&model.Asset{},
		&model.Block{},
		&model.BlockData{},
		&model.Coin{},
		&model.Delegate{},
		&model.Staking{},
		&model.Statistics{},
		&model.System{},
		&model.Transfer{},
		&model.TX{},
		&model.Validator{},
	)
	return
}

// setDefault 设置默认值
func setDefault() (err error) {
	if singleton.SystemModel, err = service.NewSystem().Initialize(singleton.DB); nil != err {
		return
	}
	if err = service.NewStatistics().Initialize(singleton.DB); nil != err {
		return
	}
	return
}

// searchIndex 建立索引
func searchIndex() (err error) {
	// 建立交易索引
	{
		var hashList []string
		if hashList, err = service.NewTX().ListHash(singleton.DB, 0, math.MaxInt64 /*TODO 采用多次加载避免内存使用过大*/); nil != err {
			return
		}
		for _, hash := range hashList {
			singleton.TXTrieTree.Add(hash, nil)
		}
	}
	// 建立高度索引
	{
		var heightList []int64
		if heightList, err = service.NewBlock().ListHeight(singleton.DB, 0, math.MaxInt64); nil != err {
			return
		}
		for _, height := range heightList {
			singleton.HeightTrieTree.Add(strconv.FormatInt(height, 10), nil)
		}
	}
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
		if err = connect(dsn, retryTimes); nil != err {
			return
		}
		if err = migrate(); nil != err {
			return
		}
		if err = setDefault(); nil != err {
			return
		}
		if err = searchIndex(); nil != err {
			return
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
