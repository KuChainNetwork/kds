package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	"kds/dbmodel"
)

// onSupplyMessages 供应消息
func (object *Analyser) onSupplyMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *dbmodel.TX) (err error) {
	return
}
