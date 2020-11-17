package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	slashingTypes "github.com/KuChainNetwork/kuchain/x/slashing/types"

	"kds/config"
	"kds/db/model"
	"kds/db/service"
	"kds/util"
)

const (
	slashMsgTypeUnJail = "unjail" // 出狱
)

// onSlashingMessages 惩罚消息
func (object *Analyser) onSlashingMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	switch msg.Type() {
	case slashMsgTypeUnJail:
		message := msg.(slashingTypes.KuMsgUnjail)
		var messageData slashingTypes.MsgUnjail
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = slashingTypes.ModuleName
		tx.Amount = util.Coin2Decimal(message.GetTransfers()[0].Amount[0], config.Exp).String() //FIXME
		tx.Denom = message.GetTransfers()[0].Amount[0].Denom
		object.fillMessageAndMessageData(tx, message, &messageData)
		if 0 == txResult.Code {
			// 出狱
			err = service.NewValidator().UpdateJailed(object.db, tx.From, false)
		}

	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
