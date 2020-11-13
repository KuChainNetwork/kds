package txAnalyser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/glog"
	abci "github.com/tendermint/tendermint/abci/types"
	"gorm.io/gorm"

	"kds/db/model"
)

const (
	evidenceMsgTypeSubmitEvidence = "submit_evidence"
	evidenceMsgTypeEquivocation   = "equivocation"
)

// onEvidenceMessages 处理证据消息
func (object *Analyser) onEvidenceMessages(db *gorm.DB,
	msg sdk.Msg,
	txResult *abci.ResponseDeliverTx,
	tx *model.TX) (err error) {
	switch msg.Type() {
	case evidenceMsgTypeSubmitEvidence:
	case evidenceMsgTypeEquivocation:
	default:
		glog.Fatalln("unknown msg type:", msg.Type())
	}
	return
}
