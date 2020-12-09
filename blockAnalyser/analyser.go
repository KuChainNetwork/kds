package blockAnalyser

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"gorm.io/gorm"

	"kds/config"
	"kds/dbmodel"
	"kds/dbservice"
	"kds/singleton"
)

// Analyser 分析器
type Analyser struct {
	db              *gorm.DB              // 数据库
	cdc             *amino.Codec          // 解码器
	newDataNotifyCh chan struct{}         // 新数据通知通道
	wg              sync.WaitGroup        // 等待组
	srvBlock        *dbservice.Block      // 区块服务
	srvStatistics   *dbservice.Statistics // 统计服务
	srvBlockData    *dbservice.BlockData  // 区块数据
}

// New 工厂方法
func New(db *gorm.DB,
	cdc *amino.Codec,
	newDataNotifyCh chan struct{}) *Analyser {
	return &Analyser{
		db:              db,
		cdc:             cdc,
		newDataNotifyCh: newDataNotifyCh,
		srvBlock:        dbservice.NewBlock(),
		srvStatistics:   dbservice.NewStatistics(),
		srvBlockData:    dbservice.NewBlockData(),
	}
}

// analyze 分析
func (object *Analyser) analyze(heightStep int64) (err error) {
	var start int64 = config.StartBlockHeight
	var blockModel *dbmodel.Block
	if blockModel, err = object.srvBlock.Latest(object.db); nil != err {
		return
	}
	if nil != blockModel {
		start = blockModel.Height + 1
	}
	for {
		var blockDataModelList []*dbmodel.BlockData
		for begin := start; begin <= singleton.LastBlockHeight; begin += heightStep {
			if blockDataModelList, err = object.srvBlockData.List(object.db,
				start,
				start+heightStep,
				false); nil != err || 0 >= len(blockDataModelList) {
				return
			}
			if nil != blockDataModelList && 0 < len(blockDataModelList) {
				break
			}
		}
		if 0 >= len(blockDataModelList) {
			break
		}
		var resultBlock ctypes.ResultBlock
		blockModelList := make([]*dbmodel.Block, 0, len(blockDataModelList))
		for i := 0; i < len(blockDataModelList); i++ {
			object.cdc.MustUnmarshalJSON(blockDataModelList[i].Block, &resultBlock)
			blockModelList = append(blockModelList, &dbmodel.Block{
				Height:    resultBlock.Block.Height,
				Hash:      resultBlock.BlockID.Hash.String(),
				Txn:       int64(len(resultBlock.Block.Txs)),
				Validator: resultBlock.Block.ValidatorsHash.String(),
				Time:      resultBlock.Block.Time,
			})
			// 索引高度
			singleton.HeightTrieTree.Add(strconv.FormatInt(resultBlock.Block.Height, 10), nil)
		}
		sort.Sort(dbmodel.NewBlockListSorter(blockModelList))
		err = object.db.Transaction(func(tx *gorm.DB) (err error) {
			if err = object.srvBlock.AddAll(tx, blockModelList); nil != err {
				return
			}
			if err = object.srvStatistics.Updates(tx, map[string]interface{}{
				"latest_height": blockModelList[len(blockModelList)-1].Height,
			}); nil != err {
				return
			}
			return
		})
		if blockModel, err = object.srvBlock.Latest(object.db); nil != err {
			return
		}
		if nil != blockModel {
			start = blockModel.Height + 1
		}
	}
	return
}

// Start 开始
func (object *Analyser) Start(heightStep int64) (err error) {
	if err = object.analyze(heightStep); nil != err {
		return
	}
	object.wg.Add(1)
	go func() {
		defer object.wg.Done()
	loop:
		for {
			select {
			case _, ok := <-object.newDataNotifyCh:
				if !ok {
					break loop
				}
				err = object.analyze(heightStep)
				if nil != err {
					glog.Errorln(err)
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()
	return
}

// Stop 停止
func (object *Analyser) Stop() (err error) {
	object.wg.Wait()
	return
}
