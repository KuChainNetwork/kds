package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	accountTypes "github.com/KuChainNetwork/kuchain/x/account/types"

	"kds/dbmodel"
	"kds/dbservice"
)

const (
	accountMsgTypeCreate     = "create@account"
	accountMsgTypeUpdateAuth = "updateauth"
)

// onAccountMessages 处理账户消息
func (object *Analyser) onAccountMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *dbmodel.TX) (err error) {
	switch msg.Type() {
	case accountMsgTypeCreate:
		message := msg.(*accountTypes.MsgCreateAccount)
		var messageData accountTypes.MsgCreateAccountData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = messageData.Name.String()
		tx.RealTo = tx.To
		if 0 == txResult.Code {
			if err = dbservice.NewAccount().Add(db, &dbmodel.Account{
				Height:  tx.Height,
				TXHash:  tx.Hash,
				Creator: messageData.Creator.String(),
				Name:    messageData.Name.String(),
				Auth:    messageData.Auth.String(),
				Time:    tx.Time,
			}); nil != err {
				glog.Fatalln(err)
				return
			}
			if err = dbservice.NewStatistics().Increment(db, "total_account", 1); nil != err {
				glog.Fatalln(err)
				return
			}
		}
		object.fillMessageAndMessageData(tx, message, messageData)

	case accountMsgTypeUpdateAuth:
		message := msg.(*accountTypes.MsgUpdateAccountAuth)
		var messageData accountTypes.MsgUpdateAccountAuthData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = messageData.Name.String()
		if 0 == txResult.Code {
			if err = dbservice.NewAccount().UpdateAuth(db, &dbmodel.Account{
				Name: messageData.Name.String(),
				Auth: messageData.Auth.String(),
			}); nil != err {
				glog.Fatalln(err)
				return
			}
		}
		object.fillMessageAndMessageData(tx, message, messageData)

	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
