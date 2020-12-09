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
	"kds/dbmodel"
	"kds/dbservice"
	"kds/singleton"
)

// Analyser 分析器
type Analyser struct {
	db              *gorm.DB              // 数据库
	cdc             *amino.Codec          // 编解码器
	newDataNotifyCh chan struct{}         // 新数据通知通道
	wg              sync.WaitGroup        // 等待组
	srvTx           *dbservice.TX         // tx数据服务
	srvStatistics   *dbservice.Statistics // 统计服务
	srvBlockData    *dbservice.BlockData  // 区块数据服务
	handlerMap      map[string]func(db *gorm.DB,
		msg sdk.Msg,
		txResult *abci.ResponseDeliverTx,
		tx *dbmodel.TX) (err error) // 处理器映射
}

// New 工厂方法
func New(db *gorm.DB,
	cdc *amino.Codec,
	newDataNotifyCh chan struct{}) *Analyser {
	object := &Analyser{
		db:              db,
		cdc:             cdc,
		newDataNotifyCh: newDataNotifyCh,
		srvTx:           dbservice.NewTX(),
		srvStatistics:   dbservice.NewStatistics(),
		srvBlockData:    dbservice.NewBlockData(),
	}
	object.handlerMap = map[string]func(db *gorm.DB,
		msg sdk.Msg,
		txResult *abci.ResponseDeliverTx,
		tx *dbmodel.TX) (err error){
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
func (object *Analyser) fillMessageAndMessageData(tx *dbmodel.TX, message, messageData interface{}) {
	tx.Message = object.cdc.MustMarshalJSON(message)
	tx.MessageData = object.cdc.MustMarshalJSON(messageData)
}

// analyze 分析数据
func (object *Analyser) analyze(heightStep int64) (err error) {
	var start int64 = config.StartBlockHeight
	var txModel *dbmodel.TX
	if txModel, err = object.srvTx.Latest(object.db); nil != err {
		return
	}
	if nil != txModel {
		start = txModel.Height + 1
	}
	for {
		var blockDataModelList []*dbmodel.BlockData
		for begin := start; begin <= singleton.LastBlockHeight; begin += heightStep {
			if blockDataModelList, err = object.srvBlockData.List(object.db,
				begin,
				begin+heightStep,
				true); nil != err {
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
		var resultBlockResults ctypes.ResultBlockResults
		var stdTx types.StdTx
		for index := 0; index < len(blockDataModelList); index++ {
			object.cdc.MustUnmarshalJSON(blockDataModelList[index].Block, &resultBlock)
			object.cdc.MustUnmarshalJSON(blockDataModelList[index].Results, &resultBlockResults)
			for j, blockTx := range resultBlock.Block.Txs {
				object.cdc.MustUnmarshalBinaryLengthPrefixed(blockTx, &stdTx)
				if 0 >= len(stdTx.Msgs) {
					continue
				}
				txResult := resultBlockResults.TxsResults[j]
				var logs []json.RawMessage
				if 0 == txResult.Code {
					singleton.Cdc.MustUnmarshalJSON([]byte(txResult.Log), &logs)
				}
				allMsg := make([]sdk.Msg, 0, len(stdTx.Msgs))
				allTx := make([]*dbmodel.TX, 0, len(stdTx.Msgs))
				for k := 0; k < len(stdTx.Msgs); k++ {
					msg := stdTx.Msgs[k]
					aTx := &dbmodel.TX{
						Hash:   fmt.Sprintf("%X", blockTx.Hash()),
						Height: resultBlock.Block.Height,
						Route:  msg.Route(),
						Type:   msg.Type(),
						Time:   resultBlock.Block.Time,
						Code:   txResult.Code,
						Log: func() string {
							if 0 == txResult.Code {
								return "[" + string(logs[k]) + "]"
							}
							return txResult.Log
						}(),
						Info:        txResult.Info,
						GasWanted:   txResult.GasWanted,
						GasUsed:     txResult.GasUsed,
						Events:      object.cdc.MustMarshalJSON(txResult.Events),
						Message:     []byte(`{}`),
						MessageData: []byte(`{}`),
					}
					allMsg = append(allMsg, msg)
					allTx = append(allTx, aTx)
					// 索引交易
					singleton.TXTrieTree.Add(aTx.Hash, nil)
				}
				err = object.db.Transaction(func(tx *gorm.DB) (err error) {
					for i := 0; i < len(stdTx.Msgs); i++ {
						msg := allMsg[i]
						aTx := allTx[i]
						if h, ok := object.handlerMap[msg.Route()]; ok {
							if err = h(tx, msg, txResult, aTx); nil != err {
								return
							}
						} else {
							glog.Fatalln("unknown route:", msg.Route())
						}
					}
					if err = object.srvTx.AddAll(tx, allTx); nil != err {
						return
					}
					if err = object.srvStatistics.Increment(tx, "total_tx", len(allTx)); nil != err {
						return
					}
					return
				})
				if nil != err {
					return
				}
			}
		}
		if txModel, err = object.srvTx.Latest(object.db); nil != err {
			return
		}
		if nil != txModel {
			start = txModel.Height + 1
		}
	}
	return
}

// Start 开始分析
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

// Stop 停止分析
func (object *Analyser) Stop() (err error) {
	object.wg.Wait()
	return
}
