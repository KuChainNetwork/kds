package blockDataGetter

import (
	"sync"
	"sync/atomic"

	clientContext "github.com/cosmos/cosmos-sdk/client/context"
	goAmino "github.com/tendermint/go-amino"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"kds/types"
)

// Getter 区块获取器
type Getter struct {
	chainId string           // 链ID
	nodeURI string           // 节点URI
	cdc     *goAmino.Codec   // 编解码器
	wg      *sync.WaitGroup  // 等待组
	node    rpcclient.Client // 客户端
	started int32            // 开始标志
}

// NewGetter 工厂方法
func NewGetter(chainId, nodeURI string,
	cdc *goAmino.Codec,
	wg *sync.WaitGroup) (object *Getter, err error) {
	object = &Getter{
		chainId: chainId,
		nodeURI: nodeURI,
		cdc:     cdc,
		wg:      wg,
	}
	if object.node, err = clientContext.NewCLIContext().
		WithChainID(object.chainId).
		WithCodec(object.cdc).
		WithNodeURI(object.nodeURI).
		WithTrustNode(true).GetNode(); nil != err {
		return
	}
	return
}

// LatestBlockHeight 最后的区块高度
func (object *Getter) LatestBlockHeight() (resultBlock *ctypes.ResultBlock, err error) {
	if nil == object.node {
		return
	}
	resultBlock, err = object.node.Block(nil)
	if nil != err {
		return
	}
	return
}

// Start 开始获取
func (object *Getter) Start(heightCh chan int64, modelCh chan *types.BlockResponse) (err error) {
	//TODO enable if subscribe
	//if err = object.blockDataGetter.Start(); nil != err {
	//	return
	//}
	//atomic.StoreInt32(&object.started, 1)
	object.wg.Add(1)
	go func() {
		defer object.wg.Done()
		var block *ctypes.ResultBlock
		var results *ctypes.ResultBlockResults
		for {
			height, ok := <-heightCh
			if !ok {
				break
			}
			if block, err = object.node.Block(&height); nil == err {
				results, err = object.node.BlockResults(&height)
			}
			modelCh <- &types.BlockResponse{
				Error:   err,
				Height:  height,
				Block:   block,
				Results: results,
			}
		}
	}()
	return
}

// Stop 停止获取
func (object *Getter) Stop() (err error) {
	if nil != object.node {
		if 1 == atomic.LoadInt32(&object.started) {
			err = object.node.Stop()
		}
	}
	return
}
