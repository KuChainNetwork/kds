package txAnalyser

import (
	"github.com/KuChainNetwork/kuchain/x/dex"
	dexTypes "github.com/KuChainNetwork/kuchain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	"kds/config"
	"kds/db/model"
	"kds/util"
)

const (
	dexMsgTypeCreateDex            = "create@dex"
	dexMsgTypeUpdateDexDescription = "updatedesc@dex"
	dexMsgTypeDestroyDex           = "destroy@dex"
	dexMsgTypeCreateSymbol         = "create@symbol"
	dexMsgTypeUpdateSymbol         = "update@symbol"
	dexMsgTypePauseSymbol          = "pause@symbol"
	dexMsgTypeRestoreSymbol        = "restore@symbol"
	dexMsgTypeShutdownSymbol       = "shutdown@symbol"
	dexMsgTypeSigIn                = "sigin"
	dexMsgTypeSigOut               = "sigout"
	dexMsgTypeDeal                 = "deal"
)

// onDexMessages 处理Dex消息
func (object *Analyser) onDexMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	switch msg.Type() {
	case dexMsgTypeCreateDex:
		message := msg.(*dexTypes.MsgCreateDex)
		var messageData dexTypes.MsgCreateDexData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = dex.ModuleName
		// only main token enable
		tx.Amount = util.Coin2Decimal(messageData.Stakings[0], config.Exp).String()
		tx.Denom = messageData.Stakings[0].Denom
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeUpdateDexDescription:
		message := msg.(*dexTypes.MsgUpdateDexDescription)
		var messageData dexTypes.MsgUpdateDexDescriptionData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeDestroyDex:
		message := msg.(*dexTypes.MsgDestroyDex)
		var messageData dexTypes.MsgDestroyDexData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeCreateSymbol:
		message := msg.(*dexTypes.MsgCreateSymbol)
		var messageData dexTypes.MsgCreateSymbolData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeUpdateSymbol:
		message := msg.(*dexTypes.MsgUpdateSymbol)
		var messageData dexTypes.MsgUpdateSymbolData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypePauseSymbol:
		message := msg.(*dexTypes.MsgPauseSymbol)
		var messageData dexTypes.MsgPauseSymbolData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeRestoreSymbol:
		message := msg.(*dexTypes.MsgRestoreSymbol)
		var messageData dexTypes.MsgRestoreSymbolData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeShutdownSymbol:
		message := msg.(*dexTypes.MsgShutdownSymbol)
		var messageData dexTypes.MsgShutdownSymbolData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeSigIn:
		message := msg.(*dexTypes.MsgDexSigIn)
		var messageData dexTypes.MsgDexSigInData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.Amount = util.Coin2Decimal(messageData.Amount[0], config.Exp).String()
		tx.Denom = messageData.Amount[0].Denom
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeSigOut:
		message := msg.(*dexTypes.MsgDexSigOut)
		var messageData dexTypes.MsgDexSigOutData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		//TODO process amount
		object.fillMessageAndMessageData(tx, message, messageData)

	case dexMsgTypeDeal:
		message := msg.(*dexTypes.MsgDexDeal)
		var messageData dexTypes.MsgDexDealData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		object.fillMessageAndMessageData(tx, message, messageData)

	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
