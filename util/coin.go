package util

import (
	"math"

	"github.com/KuChainNetwork/kuchain/chain/types"
	"github.com/shopspring/decimal"
)

// Coin2Decimal 币转换为十进制数
func Coin2Decimal(coin types.Coin, exp int) decimal.Decimal {
	dec, err := decimal.NewFromString(coin.Amount.String())
	if nil != err {
		panic(err)
	}
	return dec.DivRound(decimal.NewFromFloat(math.Pow10(exp)), int32(exp))
}
