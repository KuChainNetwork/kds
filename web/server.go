package web

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"unicode"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang/glog"
	"gorm.io/gorm"
	"kds/dbmodel"
	"kds/dbservice"
	_ "kds/docs"
	"kds/singleton"
	"kds/types"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	port          int                   // 端口
	db            *gorm.DB              // 数据库
	ln            net.Listener          // 侦听器
	wg            sync.WaitGroup        // 等待组
	startFlag     int32                 // 开始标志
	stopFlag      int32                 // 停止标志
	srvBlock      *dbservice.Block      // block数据服务
	srvTx         *dbservice.TX         // tx数据服务
	srvValidator  *dbservice.Validator  // validator数据服务
	srvDelegate   *dbservice.Delegate   // delegate数据服务
	srvCoin       *dbservice.Coin       // coin数据服务
	srvStatistics *dbservice.Statistics // Statistics数据服务
	srvAccount    *dbservice.Account    // Account数据服务
}

// NewHTTPServer 工厂方法
func NewHTTPServer(port int, db *gorm.DB) *HTTPServer {
	return &HTTPServer{
		port:          port,
		db:            db,
		srvBlock:      dbservice.NewBlock(),
		srvTx:         dbservice.NewTX(),
		srvValidator:  dbservice.NewValidator(),
		srvDelegate:   dbservice.NewDelegate(),
		srvCoin:       dbservice.NewCoin(),
		srvStatistics: dbservice.NewStatistics(),
		srvAccount:    dbservice.NewAccount(),
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
// @ID search
// @Summary 聚合搜索
// @Tags search
// @Accept json
// @Produce json
// @Router /api/v1/search/{word}/{max} [GET]
// @Param word path string true "搜索关键字"
// @Param max path int true "响应列表最大长度"
// @Success 200 {object} SearchResponse
func (object *HTTPServer) search(ctx *fiber.Ctx) (err error) {
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
		ctx.JSON(&SearchResponse{
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
	ctx.JSON(&SearchResponse{
		AddressList: addressList,                                               // 搜索地址
		TXList:      singleton.TXTrieTree.StartWith(word, maxResult, txFilter), // 搜索交易
		CoinList:    coinList,                                                  // 搜索代币
	})
	return
}

// homePageStatistics 主页统计
// @ID homePageStatistics
// @Summary 主页统计
// @Tags statistics
// @Accept json
// @Produce json
// @Router /api/v1/statistics/homePage [GET]
// @Success 200 {object} StatisticsResponse
func (object *HTTPServer) homePageStatistics(ctx *fiber.Ctx) (err error) {
	var statistics *dbmodel.Statistics
	if statistics, err = object.srvStatistics.Load(object.db); nil != err {
		return
	}
	ctx.JSON(&StatisticsResponse{
		TotalValidator: statistics.TotalValidator,
		LatestHeight:   statistics.LatestHeight,
		TotalTX:        statistics.TotalTX,
		TotalAccount:   statistics.TotalAccount,
	})
	return
}

// blockList 区块列表
// @ID blockList
// @Summary 区块列表
// @Tags block
// @Accept json
// @Produce json
// @Router /api/v1/block/list/{page_size}/{page} [GET]
// @Param page_size path int true "页大小"
// @Param page path int true "页索引"
// @Success 200 {object} BlockListResponse
func (object *HTTPServer) blockList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*dbmodel.Block
	if total, list, err = object.srvBlock.List(object.db,
		req.Offset(),
		req.Limit(),
		-1 /*height desc*/); nil != err {
		return
	}
	ctx.JSON(&BlockListResponse{
		Total: total,
		List: func() []*BlockResponse {
			brList := make([]*BlockResponse, len(list), len(list))
			for i, e := range list {
				brList[i] = &BlockResponse{
					Height:    e.Height,
					Hash:      e.Hash,
					Txn:       e.Txn,
					Validator: e.Validator,
					Time:      e.Time,
				}
			}
			return brList
		}(),
	})
	return
}

// txList 交易列表
// @ID txList
// @Summary 交易列表
// @Tags tx
// @Accept json
// @Produce json
// @Router /api/v1/tx/list/{page_size}/{page} [GET]
// @Param page_size path int true "页大小"
// @Param page path int true "页索引"
// @Success 200 {object} TXListResponse
func (object HTTPServer) txList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*dbmodel.TX
	if total, list, err = object.srvTx.List(object.db,
		req.Offset(),
		req.Limit(),
		-1); nil != err {
		return
	}
	ctx.JSON(&TXListResponse{
		Total: total,
		List: func() []*TXResponse {
			txrList := make([]*TXResponse, len(list), len(list))
			for i, e := range list {
				txrList[i] = &TXResponse{
					Hash:   e.Hash,
					Height: e.Height,
					Route:  e.Route,
					Type:   e.Type,
					From:   e.From,
					To:     e.RealTo,
					Time:   e.Time,
					Amount: e.Amount,
					Denom:  e.Denom,
				}
			}
			return txrList
		}(),
	})
	return
}

// validatorList 验证人列表
// @ID validatorList
// @Summary 验证人列表
// @Tags validator
// @Accept json
// @Produce json
// @Router /api/v1/validator/list/{type}/{page_size}/{page} [GET]
// @Param type path int true "验证人类型"
// @Param page_size path int true "页大小"
// @Param page path int true "页索引"
// @Success 200 {object} ValidatorListResponse
func (object *HTTPServer) validatorList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*dbmodel.Validator
	if total, list, err = object.srvValidator.List(object.db,
		req.Status,
		req.Offset(),
		req.Limit()); nil != err {
		return
	}
	ctx.JSON(&ValidatorListResponse{
		Total: total,
		List: func() []*ValidatorResponse {
			vrList := make([]*ValidatorResponse, len(list), len(list))
			for i, e := range list {
				vrList[i] = &ValidatorResponse{
					Rank:           i + 1,
					Validator:      e.Validator,
					Delegated:      e.Delegated.String(),
					CommissionRate: e.CommissionRate,
				}
			}
			return vrList
		}(),
	})
	return
}

// delegateList 投票列表
// @ID delegateList
// @Summary 投票列表
// @Tags delegate
// @Accept json
// @Produce json
// @Router /api/v1/delegate/list/{page_size}/{page} [GET]
// @Param page_size path int true "页大小"
// @Param page path int true "页索引"
// @Success 200 {object} DelegateListResponse
func (object *HTTPServer) delegateList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*dbmodel.Delegate
	if total, list, err = object.srvDelegate.List(object.db,
		req.Offset(),
		req.Limit()); nil != err {
		return
	}
	var validatorList []*dbmodel.Validator
	_, validatorList, err = object.srvValidator.ListIdList(object.db, func() []uint {
		idList := make([]uint, len(list))
		for i, e := range list {
			idList[i] = e.ID
		}
		return idList
	}(), 0, math.MaxInt64)
	if nil != err {
		return
	}
	commissionRateLookupTable := make(map[string]uint64, len(validatorList))
	for _, e := range validatorList {
		commissionRateLookupTable[e.Validator] = e.CommissionRate
	}
	ctx.JSON(&DelegateListResponse{
		Total: total,
		List: func() []*DelegateResponse {
			drList := make([]*DelegateResponse, len(list), len(list))
			for i, e := range list {
				drList[i] = &DelegateResponse{
					Rank:           i + 1,
					Validator:      e.Validator,
					Amount:         e.Amount,
					CommissionRate: commissionRateLookupTable[e.Validator],
				}
			}
			return drList
		}(),
	})
	return
}

// coinList 代币列表
// @ID coinList
// @Summary 代币列表
// @Tags coin
// @Accept json
// @Produce json
// @Router /api/v1/coin/list/{page_size}/{page} [GET]
// @Param page_size path int true "页大小"
// @Param page path int true "页索引"
// @Success 200 {object} CoinListResponse
func (object *HTTPServer) coinList(ctx *fiber.Ctx) (err error) {
	var req *types.PageRequest
	if req, err = object.pageRequestParam(ctx); nil != err {
		return
	}
	var total int
	var list []*dbmodel.Coin
	if total, list, err = object.srvCoin.List(object.db,
		req.Offset(),
		req.Limit()); nil != err {
		return
	}
	ctx.JSON(&CoinListResponse{
		Total: total,
		List: func() []*CoinResponse {
			crList := make([]*CoinResponse, len(list), len(list))
			for i, e := range list {
				crList[i] = &CoinResponse{
					Creator:         e.Creator,
					Symbol:          e.Symbol,
					MaxSupplyAmount: e.MaxSupplyAmount,
					IssueAmount:     e.IssueAmount,
					Description:     e.Description,
					Time:            e.Time,
				}
			}
			return crList
		}(),
	})
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
	app.Get("/swagger/*", swagger.Handler)
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
