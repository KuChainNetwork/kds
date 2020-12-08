package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"
	"kds/config"
	"kds/util"

	slashingTypes "github.com/KuChainNetwork/kuchain/x/slashing/types"

	"kds/dbmodel"
	"kds/dbservice"
)

const (
	slashMsgTypeUnJail = "unjail" // 出狱
)

// onSlashingMessages 惩罚消息
func (object *Analyser) onSlashingMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *dbmodel.TX) (err error) {
	switch msg.Type() {
	case slashMsgTypeUnJail:
		message := msg.(slashingTypes.KuMsgUnjail)
		var messageData slashingTypes.MsgUnjail
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = slashingTypes.ModuleName
		if message.Transfers != nil && len(message.Transfers) > 0 {
			tx.Amount = util.Coin2Decimal(message.GetTransfers()[0].Amount[0], config.Exp).String() //FIXME
			tx.Denom = message.GetTransfers()[0].Amount[0].Denom
		}
		object.fillMessageAndMessageData(tx, message, &messageData)
		if 0 == txResult.Code {
			// 出狱
			err = dbservice.NewValidator().UpdateJailed(object.db, tx.From, false)
		}

	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
