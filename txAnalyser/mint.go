package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	"kds/db/model"
)

// onMintMessages 处理矿工消息
func (object *Analyser) onMintMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	return
}
