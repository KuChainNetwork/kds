package types

import (
	"encoding/json"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// BlockResponse 区块响应
type BlockResponse struct {
	Error   error                      `json:"error"`
	Height  int64                      `json:"height"`
	Block   *ctypes.ResultBlock        `json:"block"`
	Results *ctypes.ResultBlockResults `json:"results"`
}

// String 字符串描述
func (object *BlockResponse) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
