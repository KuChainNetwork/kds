package blockDataGetter

import (
	"context"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"gorm.io/gorm"

	"kds/db/model"
	"kds/db/service"
	"kds/types"
)

// GetterGroup 获取组
type GetterGroup struct {
	chainId      string             // 链ID
	nodeURI      string             // 节点URI
	maxGetters   int                // 最大获取数
	db           *gorm.DB           // 数据库对象
	cdc          *amino.Codec       // 编解码器
	wg           sync.WaitGroup     // 等待组
	getters      []*Getter          // 获取器数组
	ctx          context.Context    // 上下文
	cancel       context.CancelFunc // 取消方法
	srvSystem    *service.System    // system数据服务
	srvBlockData *service.BlockData // block data数据服务
}

// NewGetterGroup 工厂方法
func NewGetterGroup(chainId, nodeURI string,
	cdc *amino.Codec,
	db *gorm.DB,
	maxWorkers int) *GetterGroup {
	object := &GetterGroup{
		chainId:      chainId,
		nodeURI:      nodeURI,
		db:           db,
		cdc:          cdc,
		maxGetters:   maxWorkers,
		getters:      make([]*Getter, maxWorkers),
		srvSystem:    service.NewSystem(),
		srvBlockData: service.NewBlockData(),
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	return object
}

// Start 启动
func (object *GetterGroup) Start(newDataNotifyCh chan struct{}) (err error) {
	blockHeightCh := make(chan int64, object.maxGetters)
	blockResultCh := make(chan *types.BlockResponse, object.maxGetters)
	for i := 0; i < object.maxGetters; i++ {
		var getter *Getter
		if getter, err = NewGetter(object.chainId, object.nodeURI, object.cdc, &object.wg); nil != err {
			return
		}
		if err = getter.Start(blockHeightCh, blockResultCh); nil != err {
			return
		}
		object.getters[i] = getter
	}
	go func() {
		defer close(blockHeightCh)
		var err error
		var resultBlock *ctypes.ResultBlock
	loop:
		for {
			select {
			case <-object.ctx.Done():
				break loop
			default:
				if resultBlock, err = object.getters[0].LatestBlockHeight(); nil != err {
					glog.Errorln(err)
					time.Sleep(1 * time.Second)
					continue
				}
				startHeight := object.srvSystem.GetLastBlockHeight(object.db)
				if startHeight >= resultBlock.Block.Height {
					time.Sleep(1 * time.Second)
					continue
				}
				total := 0
				for i := 0; i < object.maxGetters; i++ {
					curr := startHeight + int64(i)
					if curr > resultBlock.Block.Height {
						break
					}
					total++
					blockHeightCh <- curr
				}
				var blockDataList []*model.BlockData
				for i := 0; i < total; i++ {
					res := <-blockResultCh
					if nil != res.Error {
						glog.Error(res.Error)
						blockHeightCh <- res.Height
						select {
						case <-object.ctx.Done():
							return
						default:
						}
						continue
					}
					blockDataList = append(blockDataList, &model.BlockData{
						Height:  res.Height,
						Block:   object.cdc.MustMarshalJSON(res.Block),
						Results: object.cdc.MustMarshalJSON(res.Results),
					})
				}
				if 0 >= len(blockDataList) {
					continue
				}
				if err = object.db.Transaction(func(tx *gorm.DB) (err error) {
					if err = object.srvBlockData.AddAll(tx, blockDataList); nil != err {
						glog.Fatalln(err)
						return
					}
					if err = object.srvSystem.UpdateLastBlockHeight(tx, startHeight+int64(total)); nil != err {
						return
					}
					return nil
				}); nil != err {
					glog.Fatalln(err)
					continue
				}
				newDataNotifyCh <- struct{}{}
			}
		}
	}()
	return
}

// Stop 停止
func (object *GetterGroup) Stop() (err error) {
	object.cancel()
	object.wg.Wait()
	for _, worker := range object.getters {
		if err = worker.Stop(); nil != err {
			glog.Errorln(err)
		}
	}
	return
}
