package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	stakingExported "github.com/KuChainNetwork/kuchain/x/staking/exported"
	stakingTypes "github.com/KuChainNetwork/kuchain/x/staking/types"

	"kds/config"
	"kds/db/model"
	"kds/db/service"
	"kds/util"
)

const (
	stakingMsgTypeCreate          = "create@staking"
	stakingMsgTypeEdit            = "edit@staking"
	stakingMsgTypeDelegate        = "delegate"
	stakingMsgTypeBeginReDelegate = "beginredelegate"
	stakingMsgTypeBeginUnBonding  = "beginunbonding"
)

// onStakingMessages 处理抵押消息
func (object *Analyser) onStakingMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	switch msg.Type() {
	case stakingMsgTypeCreate:
		// 创建验证人
		message := msg.(stakingTypes.KuMsgCreateValidator)
		var messageData stakingTypes.MsgCreateValidator
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = stakingTypes.ModuleName
		if message.Transfers != nil && len(message.Transfers) > 0 {
			tx.Amount = util.Coin2Decimal(message.GetTransfers()[0].Amount[0], config.Exp).String() //FIXME
			tx.Denom = message.GetTransfers()[0].Amount[0].Denom
		}
		object.fillMessageAndMessageData(tx, message, &messageData)
		if 0 == txResult.Code {
			err = service.NewValidator().Add(object.db, &model.Validator{
				Height:         tx.Height,
				Validator:      messageData.ValidatorAccount.String(),
				Status:         int(stakingExported.Unbonded),
				CommissionRate: messageData.CommissionRates.Uint64(),
				Time:           tx.Time,
			})
			if nil == err {
				err = service.NewStatistics().Increment(object.db, "total_validator", 1)
			}
		}

	case stakingMsgTypeEdit:
		// 编辑验证人
		message := msg.(*stakingTypes.KuMsgEditValidator)
		var messageData stakingTypes.MsgEditValidator
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = messageData.ValidatorAccount.String()
		object.fillMessageAndMessageData(tx, message, &messageData)

	case stakingMsgTypeDelegate:
		// 抵押
		message := msg.(stakingTypes.KuMsgDelegate)
		var messageData stakingTypes.MsgDelegate
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.Sender().String()
		tx.To = messageData.DelegatorAccount.String()
		tx.Amount = util.Coin2Decimal(messageData.Amount, config.Exp).String()
		tx.Denom = messageData.Amount.Denom
		object.fillMessageAndMessageData(tx, message, &messageData)
		if 0 == txResult.Code {
			err = service.NewDelegate().Add(object.db, &model.Delegate{
				Height:    tx.Height,
				TXHash:    tx.Hash,
				Delegator: messageData.DelegatorAccount.String(),
				Validator: messageData.ValidatorAccount.String(),
				Amount:    util.Coin2Decimal(messageData.Amount, config.Exp),
				Denom:     messageData.Amount.Denom,
				Time:      tx.Time,
			})
			//TODO end block update validator status
		}

	case stakingMsgTypeBeginReDelegate:
		// 重抵押
		message := msg.(*stakingTypes.KuMsgRedelegate)
		var messageData stakingTypes.MsgBeginRedelegate
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = messageData.ValidatorDstAccount.String()
		tx.To = messageData.DelegatorAccount.String()
		tx.Amount = util.Coin2Decimal(messageData.Amount, config.Exp).String()
		tx.Denom = messageData.Amount.Denom
		object.fillMessageAndMessageData(tx, message, &messageData)
		if 0 == txResult.Code {
			err = service.NewDelegate().AddAll(object.db, []*model.Delegate{{
				Height:    tx.Height,
				TXHash:    tx.Hash,
				Delegator: messageData.DelegatorAccount.String(),
				Validator: messageData.ValidatorSrcAccount.String(),
				Amount:    util.Coin2Decimal(messageData.Amount, config.Exp).Neg(),
				Denom:     messageData.Amount.Denom,
				Time:      tx.Time,
			}, {
				Height:    tx.Height,
				TXHash:    tx.Hash,
				Delegator: messageData.DelegatorAccount.String(),
				Validator: messageData.ValidatorDstAccount.String(),
				Amount:    util.Coin2Decimal(messageData.Amount, config.Exp),
				Denom:     messageData.Amount.Denom,
				Time:      tx.Time,
			}})
		}

	case stakingMsgTypeBeginUnBonding:
		// 解除抵押
		message := msg.(*stakingTypes.KuMsgUnbond)
		var messageData stakingTypes.MsgUndelegate
		object.cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &messageData)
		tx.From = stakingTypes.ModuleName
		tx.To = messageData.ValidatorAccount.String()
		tx.Amount = util.Coin2Decimal(messageData.Amount, config.Exp).String()
		tx.Denom = messageData.Amount.Denom
		object.fillMessageAndMessageData(tx, message, &messageData)
		//TODO end block update validator status

	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
