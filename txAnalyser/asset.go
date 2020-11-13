package txAnalyser

import (
	"crypto/sha256"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	assetTypes "github.com/KuChainNetwork/kuchain/x/asset/types"

	"kds/config"
	"kds/db/model"
	"kds/db/service"
	"kds/util"
)

const (
	assetMsgTypeCreateCoin = "create"
	assetMsgTypeIssue      = "issue"
	assetMsgTypeBurn       = "burn"
	assetMsgTypeLock       = "lock"
	assetMsgTypeUnlock     = "unlock"
	assetMsgTypeExercise   = "exercise"
	assetMsgTypeApprove    = "approve"
	assetMsgTypeTransfer   = "transfer"
)

// onAssetMessages 处理资产消息
func (object *Analyser) onAssetMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	switch msg.Type() {
	case assetMsgTypeCreateCoin:
		message := msg.(*assetTypes.MsgCreateCoin)
		var messageData assetTypes.MsgCreateCoinData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = assetTypes.ModuleName
		object.fillMessageAndMessageData(tx, message, messageData)
		if 0 == txResult.Code {
			err = service.NewCoin().Add(db, &model.Coin{
				Height:          tx.Height,
				TXHash:          tx.Hash,
				Creator:         messageData.Creator.String(),
				Symbol:          messageData.Symbol.String(),
				MaxSupplyAmount: util.Coin2Decimal(messageData.MaxSupply, config.Exp).String(),
				MaxSupplyDenom:  messageData.MaxSupply.Denom,
				IssueAmount:     util.Coin2Decimal(messageData.InitSupply, config.Exp).String(),
				IssueDenom:      messageData.InitSupply.Denom,
				Time:            tx.Time,
			})
		}

	case assetMsgTypeIssue:
		message := msg.(*assetTypes.MsgIssueCoin)
		var messageData assetTypes.MsgIssueCoinData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = assetTypes.ModuleName
		object.fillMessageAndMessageData(tx, message, messageData)
		if 0 == txResult.Code {
			err = service.NewCoin().UpdateIssue(db, &model.Coin{
				Height:      tx.Height,
				Creator:     messageData.Creator.String(),
				IssueAmount: util.Coin2Decimal(messageData.Amount, config.Exp).String(),
				IssueDenom:  messageData.Amount.Denom,
				Time:        tx.Time,
			})
		}

	case assetMsgTypeBurn:
		message := msg.(*assetTypes.MsgBurnCoin)
		var messageData assetTypes.MsgBurnCoinData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = assetTypes.ModuleName
		object.fillMessageAndMessageData(tx, message, messageData)

	case assetMsgTypeLock:
		message := msg.(*assetTypes.MsgLockCoin)
		var messageData assetTypes.MsgLockCoinData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = assetTypes.ModuleName
		object.fillMessageAndMessageData(tx, message, messageData)

	case assetMsgTypeUnlock:
		message := msg.(*assetTypes.MsgUnlockCoin)
		var messageData assetTypes.MsgUnlockCoinData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = assetTypes.ModuleName
		object.fillMessageAndMessageData(tx, message, messageData)

	case assetMsgTypeExercise:
		message := msg.(*assetTypes.MsgExerciseCoin)
		var messageData assetTypes.MsgExerciseCoinData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = assetTypes.ModuleName
		object.fillMessageAndMessageData(tx, message, messageData)

	case assetMsgTypeApprove:
		message := msg.(*assetTypes.MsgApprove)
		var messageData assetTypes.MsgApproveData
		if messageData, err = message.GetData(); nil != err {
			glog.Fatalln(err)
			return
		}
		tx.From = messageData.Sender().String()
		tx.To = assetTypes.ModuleName
		object.fillMessageAndMessageData(tx, message, messageData)

	case assetMsgTypeTransfer:
		message := msg.(*assetTypes.MsgTransfer)
		var tss []*model.Transfer
		hash := sha256.New()
		for _, ts := range message.GetTransfers() {
			hash.Reset()
			hash.Write(object.cdc.MustMarshalBinaryBare(msg))
			hash.Write(object.cdc.MustMarshalBinaryBare(ts))
			hashHex := hex.EncodeToString(hash.Sum(nil))
			tss = append(tss, &model.Transfer{
				TxHeight: tx.Height,
				TxHash:   tx.Hash,
				Hash:     hashHex,
				From:     ts.From.String(),
				To:       ts.To.String(),
				Time:     tx.Time,
			})
			if 0 == txResult.Code {
				tx := tss[len(tss)-1]
				//TODO coins
				tx.Amount = util.Coin2Decimal(ts.Amount[0], config.Exp).String()
				tx.Denom = ts.Amount[0].Denom
			}
		}
		err := service.NewTransfer().AddAll(db, tss)
		if nil != err {
			glog.Fatalln(err)
		}

	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
