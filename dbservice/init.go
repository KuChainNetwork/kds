package dbservice

import (
	"kds/dbmodel"
	"kds/singleton"
	"strconv"
)

const (
	indexBatchSize = 8192
)

// setDefault 设置默认值
func setDefault() (err error) {
	var m *dbmodel.System
	if m, err = NewSystem().Initialize(singleton.DB); nil != err {
		return
	}
	if err = NewStatistics().Initialize(singleton.DB); nil != err {
		return
	}
	singleton.LastBlockHeight = m.LastBlockHeight
	return
}

// searchIndex 建立索引
func searchIndex() (err error) {
	// 建立交易索引
	var hashList []string
	for i := 0; ; i++ {
		if hashList, err = NewTX().ListHash(singleton.DB, i*indexBatchSize, indexBatchSize); nil != err {
			return
		}
		for _, hash := range hashList {
			singleton.TXTrieTree.Add(hash, nil)
		}
		if indexBatchSize > len(hashList) {
			break
		}
	}
	// 建立高度索引
	var heightList []int64
	for i := 0; ; i++ {
		if heightList, err = NewBlock().ListHeight(singleton.DB, i*indexBatchSize, indexBatchSize); nil != err {
			return
		}
		for _, height := range heightList {
			singleton.HeightTrieTree.Add(strconv.FormatInt(height, 10), nil)
		}
		if indexBatchSize > len(heightList) {
			break
		}
	}
	return
}

// Initialize 初始化
func Initialize() (err error) {
	for _, fn := range []func() error{
		setDefault,
		searchIndex,
	} {
		if err = fn(); nil != err {
			break
		}
	}
	return
}
