package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	"kds/db/model"
)

const (
	distributionMsgTypeWithdrawCCCId     = "withdrawcccid"
	distributionMsgTypeWithdrawDelReward = "withdrawdelreward"
	distributionMsgTypeWithdrawValCom    = "withdrawvalcom"
)

// onDistributionMessages 处理分红消息
func (object *Analyser) onDistributionMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	switch msg.Type() {
	case distributionMsgTypeWithdrawCCCId:
	case distributionMsgTypeWithdrawDelReward:
	case distributionMsgTypeWithdrawValCom:
	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
