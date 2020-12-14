package web

import (
	"time"

	"github.com/shopspring/decimal"
)

// SearchResponse 搜索响应
type SearchResponse struct {
	AddressList []string `json:"address_list"` // 地址列表
	TXList      []string `json:"tx_list"`      // 交易列表
	HeightList  []string `json:"height_list"`  // 高度列表
	CoinList    []string `json:"coin_list"`    // 代币列表
}

// StatisticsResponse 统计响应
type StatisticsResponse struct {
	TotalValidator int `json:"total_validator"` // 总验证人
	LatestHeight   int `json:"latest_height"`   // 最后区块高度
	TotalTX        int `json:"total_tx"`        // 总交易数
	TotalAccount   int `json:"total_account"`   // 总账户数
}

// BlockResponse 区块响应
type BlockResponse struct {
	Height    int64     `json:"height"`    // 高度
	Hash      string    `json:"hash"`      // 交易HASH
	Txn       int64     `json:"txn"`       // 交易数量
	Validator string    `json:"validator"` // 验证人
	Time      time.Time `json:"time"`      // 时间
}

// BlockListResponse 区块列表响应
type BlockListResponse struct {
	Total int              `json:"total"` // 总数
	List  []*BlockResponse `json:"list"`  // 列表
}

// TXResponse 交易响应
type TXResponse struct {
	Hash   string    `json:"hash"`   // 交易HASH
	Height int64     `json:"height"` // 交易高度
	Route  string    `json:"route"`  // 交易路由
	Type   string    `json:"type"`   // 交易类型
	From   string    `json:"from"`   // 交易发起方
	To     string    `json:"to"`     // 交易接收方
	Time   time.Time `json:"time"`   // 交易时间
	Amount string    `json:"amount"` // 交易总额
	Denom  string    `json:"denom"`  // 交易类型
}

// TXListResponse 区块列表响应
type TXListResponse struct {
	Total int           `json:"total"` // 总数
	List  []*TXResponse `json:"list"`  // 列表
}

// ValidatorResponse 验证人类型
type ValidatorResponse struct {
	Rank           int    `json:"rank"`            // 排名
	Validator      string `json:"validator"`       // 验证人
	Delegated      string `json:"delegated"`       // 抵押
	CommissionRate uint64 `json:"commission_rate"` // 佣金比例
}

// ValidatorListResponse 验证人列表响应
type ValidatorListResponse struct {
	Total int                  `json:"total"` // 总数
	List  []*ValidatorResponse `json:"list"`  // 列表
}

// DelegateResponse 投票响应
type DelegateResponse struct {
	Rank           int             `json:"rank"`            // 排名
	Validator      string          `json:"validator"`       // 节点名
	Amount         decimal.Decimal `json:"amount"`          // 实时投票数
	CommissionRate uint64          `json:"commission_rate"` // 佣金比例
}

// DelegateListResponse 抵押列表响应
type DelegateListResponse struct {
	Total int                 `json:"total"` // 总数
	List  []*DelegateResponse `json:"list"`  // 列表
}

// CoinResponse 代币响应
type CoinResponse struct {
	Creator         string    `json:"creator"`           // 发行人
	Symbol          string    `json:"symbol"`            // 名称
	MaxSupplyAmount string    `json:"max_supply_amount"` // 发型总量
	IssueAmount     string    `json:"issue_amount"`      // 流通总量
	Description     string    `json:"description"`       // 描述
	Time            time.Time `json:"time"`              // 创建时间
}

// CoinListResponse 抵押列表响应
type CoinListResponse struct {
	Total int             `json:"total"` // 总数
	List  []*CoinResponse `json:"list"`  // 列表
}
