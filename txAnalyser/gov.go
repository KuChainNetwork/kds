package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	govTypes "github.com/KuChainNetwork/kuchain/x/gov/types"

	"kds/config"
	"kds/db/model"
	"kds/util"
)

const (
	govMsgTypeSubmitProposal = "submitproposal"
	govMsgTypeDeposit        = "deposit"
	govMsgTypeVote           = "vote"
)

// onGovMessages 处理治理消息
func (object *Analyser) onGovMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	switch msg.Type() {
	case govMsgTypeSubmitProposal:
		// 提议
		message := msg.(govTypes.KuMsgSubmitProposal)
		var messageData govTypes.MsgSubmitProposal
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = messageData.Proposer.String()
		object.fillMessageAndMessageData(tx, message, &messageData)

	case govMsgTypeDeposit:
		// 质押
		message := msg.(govTypes.KuMsgDeposit)
		var messageData govTypes.MsgDeposit
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = messageData.Depositor.String()
		tx.Amount = util.Coin2Decimal(messageData.Amount[0], config.Exp).String()
		tx.Denom = messageData.Amount[0].Denom
		object.fillMessageAndMessageData(tx, message, &messageData)

	case govMsgTypeVote:
		// 投票
		message := msg.(govTypes.KuMsgVote)
		var messageData govTypes.MsgVote
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = messageData.Voter.String()
		object.fillMessageAndMessageData(tx, message, &messageData)

	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
