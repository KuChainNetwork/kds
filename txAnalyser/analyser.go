package txAnalyser

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/KuChainNetwork/kuchain/chain/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	"github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"gorm.io/gorm"

	abci "github.com/tendermint/tendermint/abci/types"

	"kds/config"
	"kds/db/model"
	"kds/db/service"
	"kds/singleton"
)

// Analyser 分析器
type Analyser struct {
	db              *gorm.DB            // 数据库
	cdc             *amino.Codec        // 编解码器
	newDataNotifyCh chan struct{}       // 新数据通知通道
	wg              sync.WaitGroup      // 等待组
	srvTx           *service.TX         // tx数据服务
	srvStatistics   *service.Statistics // 统计服务
	handlerMap      map[string]func(db *gorm.DB,
		msg sdk.Msg,
		txResult *abci.ResponseDeliverTx,
		tx *model.TX) (err error) // 处理器映射
}

// New 工厂方法
func New(db *gorm.DB,
	cdc *amino.Codec,
	newDataNotifyCh chan struct{}) *Analyser {
	object := &Analyser{
		db:              db,
		cdc:             cdc,
		newDataNotifyCh: newDataNotifyCh,
		srvTx:           service.NewTX(),
		srvStatistics:   service.NewStatistics(),
	}
	object.handlerMap = map[string]func(db *gorm.DB,
		msg sdk.Msg,
		txResult *abci.ResponseDeliverTx,
		tx *model.TX) (err error){
		"account":        object.onAccountMessages,
		"asset":          object.onAssetMessages,
		"dex":            object.onDexMessages,
		"kudistribution": object.onDistributionMessages,
		"kuevidence":     object.onEvidenceMessages,
		"kugov":          object.onGovMessages,
		"mint":           object.onMintMessages,
		"kuslashing":     object.onSlashingMessages,
		"kustaking":      object.onStakingMessages,
		"supply":         object.onSupplyMessages,
	}
	return object
}

// fillMessageAndMessageData 填充message和messageData
func (object *Analyser) fillMessageAndMessageData(tx *model.TX, message, messageData interface{}) {
	tx.Message = object.cdc.MustMarshalJSON(message)
	tx.MessageData = object.cdc.MustMarshalJSON(messageData)
}

// analyze 分析数据
func (object *Analyser) analyze(limit int64) (err error) {
	start := int64(config.StartBlockHeight)
	var block *model.TX
	blockDataSrv := service.NewBlockData()
	var blockDataList []*model.BlockData
	for {
		if block, err = object.srvTx.Latest(object.db); nil != err {
			return
		}
		if nil != block {
			start = block.Height + 1
		}
		if blockDataList, err = blockDataSrv.List(object.db, start, limit, true); nil != err || 0 >= len(blockDataList) {
			return
		}
		var block ctypes.ResultBlock
		var blockResults ctypes.ResultBlockResults
		var stdTx types.StdTx
		for i := 0; i < len(blockDataList); i++ {
			object.cdc.MustUnmarshalJSON(blockDataList[i].Block, &block)
			object.cdc.MustUnmarshalJSON(blockDataList[i].Results, &blockResults)
			for j, blockTx := range block.Block.Txs {
				object.cdc.MustUnmarshalBinaryLengthPrefixed(blockTx, &stdTx)
				if 0 >= len(stdTx.Msgs) {
					continue
				}

				err = object.db.Transaction(func(tx *gorm.DB) (err error) {
					// todo:优化逻辑

					if 1 == len(stdTx.Msgs) {
						msg := stdTx.Msgs[0]
						txResult := blockResults.TxsResults[j]
						aTx := &model.TX{
							Hash:        fmt.Sprintf("%X", blockTx.Hash()),
							Height:      block.Block.Height,
							Route:       msg.Route(),
							Type:        msg.Type(),
							Time:        block.Block.Time,
							Code:        txResult.Code,
							Log:         txResult.Log,
							Info:        txResult.Info,
							GasWanted:   txResult.GasWanted,
							GasUsed:     txResult.GasUsed,
							Events:      object.cdc.MustMarshalJSON(txResult.Events),
							Message:     []byte(`{}`),
							MessageData: []byte(`{}`),
						}
						// 索引交易
						singleton.TXTrieTree.Add(aTx.Hash, nil)
						if h, ok := object.handlerMap[msg.Route()]; ok {
							if err = h(tx, msg, txResult, aTx); nil != err {
								return
							}
						} else {
							glog.Fatalln("unknown route:", msg.Route())
						}
						if err = object.srvTx.Add(tx, aTx); nil != err {
							return
						}
						if err = object.srvStatistics.Increment(tx, "total_tx", 1); nil != err {
							return
						}
					} else if 1 < len(stdTx.Msgs) {
						for k := 0; k < len(stdTx.Msgs); k++ {
							msg := stdTx.Msgs[k]
							txResult := blockResults.TxsResults[j]
							var logs []json.RawMessage
							singleton.Cdc.MustUnmarshalJSON([]byte(txResult.Log), &logs)
							aTx := &model.TX{
								Hash:        fmt.Sprintf("%X", blockTx.Hash()),
								Height:      block.Block.Height,
								Route:       msg.Route(),
								Type:        msg.Type(),
								Time:        block.Block.Time,
								Code:        txResult.Code,
								Log:         string(logs[k]),
								Info:        txResult.Info,
								GasWanted:   txResult.GasWanted,
								GasUsed:     txResult.GasUsed,
								Events:      object.cdc.MustMarshalJSON(txResult.Events),
								Message:     []byte(`{}`),
								MessageData: []byte(`{}`),
							}
							// 索引交易
							singleton.TXTrieTree.Add(aTx.Hash, nil)
							if h, ok := object.handlerMap[msg.Route()]; ok {
								if err = h(tx, msg, txResult, aTx); nil != err {
									return
								}
							} else {
								glog.Fatalln("unknown route:", msg.Route())
							}
							if err = object.srvTx.Add(tx, aTx); nil != err {
								return
							}
							if err = object.srvStatistics.Increment(tx, "total_tx", 1); nil != err {
								return
							}
						}
					}
					return
				})
				if nil != err {
					return
				}
			}
		}
	}
	return
}

// Start 开始分析
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

// Stop 停止分析
func (object *Analyser) Stop() (err error) {
	object.wg.Wait()
	return
}
