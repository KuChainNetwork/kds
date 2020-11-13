package web

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang/glog"
	"gorm.io/gorm"

	"kds/db/model"
	"kds/db/service"
	"kds/singleton"
	"kds/types"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	port          int                 // 端口
	db            *gorm.DB            // 数据库
	ln            net.Listener        // 侦听器
	wg            sync.WaitGroup      // 等待组
	startFlag     int32               // 开始标志
	stopFlag      int32               // 停止标志
	srvBlock      *service.Block      // block数据服务
	srvTx         *service.TX         // tx数据服务
	srvValidator  *service.Validator  // validator数据服务
	srvDelegate   *service.Delegate   // delegate数据服务
	srvCoin       *service.Coin       // coin数据服务
	srvStatistics *service.Statistics // Statistics数据服务
	srvAccount    *service.Account    // Account数据服务
}

// NewHTTPServer 工厂方法
func NewHTTPServer(port int, db *gorm.DB) *HTTPServer {
	return &HTTPServer{
		port:          port,
		db:            db,
		srvBlock:      service.NewBlock(),
		srvTx:         service.NewTX(),
		srvValidator:  service.NewValidator(),
		srvDelegate:   service.NewDelegate(),
		srvCoin:       service.NewCoin(),
		srvStatistics: service.NewStatistics(),
		srvAccount:    service.NewAccount(),
	}
}

// pageRequestParam 页请求参数
func (object *HTTPServer) pageRequestParam(ctx *fiber.Ctx) (req *types.PageRequest, err error) {
	var _req types.PageRequest
	if _req.PageSize, err = strconv.Atoi(ctx.Params("page_size")); nil != err {
		return
	}
	if _req.Page, err = strconv.Atoi(ctx.Params("page")); nil != err {
		return
	}
	if typeStr := ctx.Params("status", ""); 0 < len(typeStr) {
		_req.Status, err = strconv.Atoi(ctx.Params("status"))
	}
	req = &_req
	return
}

// search 搜索
func (object *HTTPServer) search(ctx *fiber.Ctx) (err error) {
	type Response struct {
		AddressList []string `json:"address_list"`
		TXList      []string `json:"tx_list"`
		HeightList  []string `json:"height_list"`
		CoinList    []string `json:"coin_list"`
	}
	// 提取参数
	word := ctx.Params("word")
	max := ctx.Params("max") // 可选
	maxResult := 5
	if 0 < len(max) {
		if maxResult, err = strconv.Atoi(max); nil != err {
			return
		}
		if 32 < maxResult {
			maxResult = 32
		}
	}
	// 检查是否全为数字
	allIsDigit := true
	for _, ch := range word {
		if !unicode.IsDigit(ch) {
			allIsDigit = false
			break
		}
	}
	// 交易过滤器
	txFilter := func(word string, list []string) bool {
		if 64 != len(word) {
			return false
		}
		for _, ch := range list {
			if ch == word {
				return false
			}
		}
		return true
	}
	// 高度过滤器
	heightFilter := func(word string, list []string) bool {
		for _, ch := range list {
			if ch == word {
				return false
			}
		}
		return true
	}
	// 全部为数字
	if allIsDigit {
		ctx.JSON(&Response{
			TXList:     singleton.TXTrieTree.StartWith(word, maxResult, txFilter),         // 搜索交易
			HeightList: singleton.HeightTrieTree.StartWith(word, maxResult, heightFilter), // 搜索高度
		})
		return
	}
	// 搜索地址
	var addressList []string
	if addressList, err = object.srvAccount.LikeAddress(object.db, word, 0, maxResult); nil != err {
		return
	}
	// 搜索代币
	var coinList []string
	if coinList, err = object.srvCoin.LikeSymbol(object.db, word, 0, maxResult); nil != err {
		return
	}
	ctx.JSON(&Response{
		AddressList: addressList,                                               // 搜索地址
		TXList:      singleton.TXTrieTree.StartWith(word, maxResult, txFilter), // 搜索交易
		CoinList:    coinList,                                                  // 搜索代币
	})
	return
}

// homePageStatistics 主页统计
func (object *HTTPServer) homePageStatistics(ctx *fiber.Ctx) (err error) {
	var statistics *model.Statistics
	if statistics, err = object.srvStatistics.Load(object.db); nil != err {
		return
	}
	ctx.JSON(statistics)
	return
}

// blockList 区块列表
func (object *HTTPServer) blockList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*model.Block
	if total, list, err = object.srvBlock.List(object.db,
		req.Offset(),
		req.Limit(),
		-1 /*height desc*/); nil != err {
		return
	}
	ctx.JSON(&types.PageResponse{
		Total: total,
		List:  list,
	})
	return
}

// txList 交易列表
func (object HTTPServer) txList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*model.TX
	if total, list, err = object.srvTx.List(object.db,
		req.Offset(),
		req.Limit(),
		-1); nil != err {
		return
	}
	ctx.JSON(&types.PageResponse{
		Total: total,
		List:  list,
	})
	return
}

// validatorList 验证人列表
func (object *HTTPServer) validatorList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*model.Validator
	if total, list, err = object.srvValidator.List(object.db,
		req.Status,
		req.Offset(),
		req.Limit()); nil != err {
		return
	}
	ctx.JSON(&types.PageResponse{
		Total: total,
		List:  list,
	})
	return
}

// delegateList 投票列表
func (object *HTTPServer) delegateList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*model.Delegate
	if total, list, err = object.srvDelegate.List(object.db,
		req.Offset(),
		req.Limit()); nil != err {
		return
	}
	ctx.JSON(&types.PageResponse{
		Total: total,
		List:  list,
	})
	return
}

// coinList 代币列表
func (object *HTTPServer) coinList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*model.Coin
	if total, list, err = object.srvCoin.List(object.db,
		req.Offset(),
		req.Limit()); nil != err {
		return
	}
	ctx.JSON(&types.PageResponse{
		Total: total,
		List:  list,
	})
	return
	return
}

// Start 开始
func (object *HTTPServer) Start() (err error) {
	if !atomic.CompareAndSwapInt32(&object.startFlag, 0, 1) {
		return
	}
	if object.ln, err = net.Listen("tcp", fmt.Sprintf(":%d", object.port)); nil != err {
		return
	}
	app := fiber.New()
	app.Use(logger.New())
	app.Get("/api/v1/search/:word/:max?", object.search)
	gpStatistics := app.Group("/api/v1/statistics")
	gpStatistics.Get("/homePage", object.homePageStatistics)
	gpBlock := app.Group("/api/v1/block")
	gpBlock.Get("/list/:page_size/:page", object.blockList)
	gpTx := app.Group("/api/v1/tx")
	gpTx.Get("/list/:page_size/:page", object.txList)
	gpValidator := app.Group("/api/v1/validator")
	gpValidator.Get("/list/:type/:page_size/:page", object.validatorList)
	gpDelegate := app.Group("/api/v1/delegate")
	gpDelegate.Get("/list/:page_size/:page", object.delegateList)
	gpCoin := app.Group("/api/v1/coin")
	gpCoin.Get("/list/:page_size/:page", object.coinList)
	object.wg.Add(1)
	go func() {
		defer object.wg.Done()
		if err = app.Listener(object.ln); nil != err {
			glog.Errorln(err)
		}
	}()
	return
}

// Stop 停止
func (object *HTTPServer) Stop() (err error) {
	if 1 != atomic.LoadInt32(&object.startFlag) {
		return
	}
	if !atomic.CompareAndSwapInt32(&object.stopFlag, 0, 1) {
		return
	}
	err = object.ln.Close()
	object.wg.Wait()
	return
}
