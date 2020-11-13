package blockAnalyser

import (
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"gorm.io/gorm"

	"kds/config"
	"kds/db/model"
	"kds/db/service"
	"kds/singleton"
)

// Analyser 分析器
type Analyser struct {
	db              *gorm.DB            // 数据库
	cdc             *amino.Codec        // 解码器
	newDataNotifyCh chan struct{}       // 新数据通知通道
	wg              sync.WaitGroup      // 等待组
	srvBlock        *service.Block      // 区块服务
	srvStatistics   *service.Statistics // 统计服务
}

// New 工厂方法
func New(db *gorm.DB,
	cdc *amino.Codec,
	newDataNotifyCh chan struct{}) *Analyser {
	return &Analyser{
		db:              db,
		cdc:             cdc,
		newDataNotifyCh: newDataNotifyCh,
		srvBlock:        service.NewBlock(),
		srvStatistics:   service.NewStatistics(),
	}
}

// analyze 分析
func (object *Analyser) analyze(limit int64) (err error) {
	start := int64(config.StartBlockHeight)
	var block *model.Block
	blockDataSrv := service.NewBlockData()
	var blockDataList []*model.BlockData
	for {
		if block, err = object.srvBlock.Latest(object.db); nil != err {
			return
		}
		if nil != block {
			start = block.Height + 1
		}
		if blockDataList, err = blockDataSrv.List(object.db, start, limit, false); nil != err || 0 >= len(blockDataList) {
			return
		}
		var block ctypes.ResultBlock
		blocks := make([]*model.Block, 0, limit)
		for i := 0; i < len(blockDataList); i++ {
			object.cdc.MustUnmarshalJSON(blockDataList[i].Block, &block)
			blocks = append(blocks, &model.Block{
				Height:    block.Block.Height,
				Hash:      block.BlockID.Hash.String(),
				Txn:       int64(len(block.Block.Txs)),
				Validator: block.Block.ValidatorsHash.String(),
				Time:      block.Block.Time,
			})
			// 索引高度
			singleton.HeightTrieTree.Add(strconv.FormatInt(block.Block.Height, 10), nil)
		}
		err = object.db.Transaction(func(tx *gorm.DB) (err error) {
			if err = object.srvBlock.AddAll(tx, blocks); nil != err {
				return
			}
			if err = object.srvStatistics.Updates(tx, map[string]interface{}{
				"latest_height": blocks[len(blocks)-1].Height,
			}); nil != err {
				return
			}
			return
		})
	}
	return
}

// Start 开始
func (object *Analyser) Start(limit int64) (err error) {
	if err = object.analyze(limit); nil != err {
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
				err = object.analyze(limit)
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
