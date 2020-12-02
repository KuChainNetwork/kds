package util

import (
	"testing"

	"github.com/KuChainNetwork/kuchain/chain/types"
	"kds/config"
)

func TestCoin2Decimal(t *testing.T) {
	t.Log(Coin2Decimal(types.NewInt64Coin("kratos/kts", 999999999999999999), config.Exp))
}
