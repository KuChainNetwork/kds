package genesis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	chainTypes "github.com/KuChainNetwork/kuchain/chain/types"
	accountTypes "github.com/KuChainNetwork/kuchain/x/account/types"
	assetTypes "github.com/KuChainNetwork/kuchain/x/asset/types"
	genUtilTypes "github.com/KuChainNetwork/kuchain/x/genutil/types"
	stakingTypes "github.com/KuChainNetwork/kuchain/x/staking/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/types"
	"gorm.io/gorm"

	"kds/config"
	"kds/db/model"
	"kds/db/service"
	"kds/singleton"
	"kds/util"
)

var (
	ErrGetGenesis   = errors.New("get genesis json error")      // 获取创世JSON错误
	ErrChainID      = errors.New("genesis json chain id error") // 创世JSON链ID不匹配
	ErrAccountState = errors.New("genesis account state error") // 账户状态错误
	ErrAssetState   = errors.New("genesis asset state error")   // 资产状态错误
)

// Genesis 创世
type Genesis struct {
	genesisURL string            // 创世JSON文件URL
	doc        *types.GenesisDoc // 创世JSON公共文件
}

// New 工厂方法
func New(genesisURL string) *Genesis {
	return &Genesis{genesisURL: genesisURL}
}

// processAccount 处理账户
func (object *Genesis) processAccount(cdc *amino.Codec,
	appState map[string]json.RawMessage,
	db *gorm.DB) (err error) {
	//genesis account
	raw := appState[accountTypes.ModuleName]
	if 0 >= len(raw) {
		err = ErrAccountState
		return
	}
	var state accountTypes.GenesisState
	if err = cdc.UnmarshalJSON(raw, &state); nil != err {
		return
	}
	var accountList []*model.Account
	for _, account := range state.Accounts {
		accountList = append(accountList, &model.Account{
			Height:    config.StartBlockHeight,
			TXHash:    "",
			Creator:   object.doc.ChainID,
			AccountID: account.GetID().String(),
			Number:    account.GetAccountNumber(),
			Name:      account.GetName().String(),
			Auth:      account.GetAuth().String(),
			Time:      object.doc.GenesisTime,
		})
	}
	if err = service.NewAccount().AddAll(db, accountList); nil != err {
		return
	}
	if err = service.NewStatistics().Increment(db, "total_account", len(accountList)); nil != err {
		return
	}
	return
}

// processAsset 处理资产
func (object *Genesis) processAsset(cdc *amino.Codec,
	appState map[string]json.RawMessage,
	db *gorm.DB) (err error) {
	//genesis asset
	raw := appState[assetTypes.ModuleName]
	if 0 >= len(raw) {
		err = ErrAssetState
		return
	}
	var state assetTypes.GenesisState
	if err = cdc.UnmarshalJSON(raw, &state); nil != err {
		return
	}
	var assetList []*model.Asset
	var coinList []*model.Coin
	for _, asset := range state.GenesisAssets {
		for _, coin := range asset.GetCoins() {
			assetList = append(assetList, &model.Asset{
				Height: config.StartBlockHeight,
				TXHash: "",
				RealId: asset.GetID().String(),
				Amount: util.Coin2Decimal(coin, config.Exp).String(),
				Denom:  coin.Denom,
				Time:   object.doc.GenesisTime,
			})
		}
	}
	for _, coin := range state.GenesisCoins {
		coinList = append(coinList, &model.Coin{
			Height:          config.StartBlockHeight,
			TXHash:          "",
			Creator:         coin.GetCreator().String(),
			Symbol:          coin.GetSymbol().String(),
			MaxSupplyAmount: util.Coin2Decimal(coin.GetMaxSupply(), config.Exp).String(),
			MaxSupplyDenom:  coin.GetMaxSupply().Denom,
			Description:     coin.GetDescription(),
			IssueAmount:     util.Coin2Decimal(coin.GetMaxSupply(), config.Exp).String(),
			IssueDenom:      coin.GetMaxSupply().Denom,
			Time:            object.doc.GenesisTime,
		})
	}
	if err = service.NewAsset().AddAll(db, assetList); nil == err {
		err = service.NewCoin().AddAll(db, coinList)
	}
	return
}

// processGenUtil 处理创世
func (object *Genesis) processGenUtil(cdc *amino.Codec,
	appState map[string]json.RawMessage,
	db *gorm.DB) (err error) {
	//genUtil asset
	raw := appState[genUtilTypes.ModuleName]
	if 0 >= len(raw) {
		err = ErrAssetState
		return
	}
	var state genUtilTypes.GenesisState
	if err = cdc.UnmarshalJSON(raw, &state); nil != err {
		return
	}
	srvValidator := service.NewValidator()
	srvTransfer := service.NewTransfer()
	srvStaking := service.NewStaking()
	srvDelegate := service.NewDelegate()
	srvStatistics := service.NewStatistics()
	var stdTx chainTypes.StdTx
	for _, raw = range state.GenTxs {
		if err = cdc.UnmarshalJSON(raw, &stdTx); nil != err {
			return
		}
		for _, msg := range stdTx.Msgs {
			switch msg.Type() {
			case "create@staking":
				message := msg.(stakingTypes.KuMsgCreateValidator)
				var create stakingTypes.MsgCreateValidator
				cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &create)
				if err = srvValidator.Add(db, &model.Validator{
					Height:         config.StartBlockHeight,
					Validator:      create.ValidatorAccount.String(),
					CommissionRate: create.CommissionRates.Uint64(),
					Time:           object.doc.GenesisTime,
				}); nil != err {
					return
				}
				if err = srvStatistics.Increment(db, "total_validator", 1); nil != err {
					return
				}

			case "delegate":
				message := msg.(stakingTypes.KuMsgDelegate)
				var delegate stakingTypes.MsgDelegate
				cdc.MustUnmarshalBinaryLengthPrefixed(message.GetData(), &delegate)
				sum256 := sha256.Sum256(raw)
				hash := sha256.New()
				var txList []*model.Transfer
				for _, tx := range message.Transfers {
					hash.Reset()
					hash.Write(cdc.MustMarshalBinaryBare(msg))
					hash.Write(cdc.MustMarshalBinaryBare(tx))
					txList = append(txList, &model.Transfer{
						TxHeight: 1,
						TxHash:   hex.EncodeToString(sum256[:]),
						Hash:     hex.EncodeToString(hash.Sum(nil)),
						Auth:     delegate.Sender().String(),
						From:     tx.From.String(),
						To:       tx.To.String(),
						Amount:   util.Coin2Decimal(tx.Amount[0], config.Exp).String(), //FIXME
						Denom:    tx.Amount[0].Denom,
						Time:     object.doc.GenesisTime,
					})
				}
				if err = srvTransfer.AddAll(db, txList); nil != err {
					return
				}
				if err = srvStaking.Add(db, &model.Staking{
					Height:        config.StartBlockHeight,
					TXHash:        "",
					Validator:     delegate.ValidatorAccount.String(),
					Delegator:     delegate.DelegatorAccount.String(),
					StakingAmount: util.Coin2Decimal(delegate.Amount, config.Exp).String(),
					StakingDenom:  delegate.Amount.Denom,
					Time:          object.doc.GenesisTime,
				}); nil != err {
					return
				}
				if err = srvDelegate.Add(db, &model.Delegate{
					Height:    config.StartBlockHeight,
					TXHash:    "",
					Validator: delegate.ValidatorAccount.String(),
					Delegator: delegate.DelegatorAccount.String(),
					Amount:    util.Coin2Decimal(delegate.Amount, config.Exp),
					Denom:     delegate.Amount.Denom,
					Time:      object.doc.GenesisTime,
				}); nil != err {
					return
				}
			}
		}
	}
	return
}

// Initialize 初始化
func (object *Genesis) Initialize(db *gorm.DB,
	cdc *amino.Codec,
	blockHeight int64) (err error, done bool) {
	if 1 <= blockHeight {
		return
	}
	done = true
	var res *http.Response
	if res, err = http.Get(object.genesisURL); nil != err {
		return
	}
	if nil == res.Body || 200 > res.StatusCode || 300 <= res.StatusCode {
		err = ErrGetGenesis
		return
	}
	var raw json.RawMessage
	if func() (err error) {
		defer res.Body.Close()
		if raw, err = ioutil.ReadAll(res.Body); nil == err {
			type Res struct {
				Result struct {
					Genesis json.RawMessage `json:"genesis"`
				} `json:"result"`
			}
			var res Res
			if err = json.Unmarshal(raw, &res); nil == err {
				raw = res.Result.Genesis
			}
		}
		return err
	}(); nil != err {
		return
	}
	if object.doc, err = types.GenesisDocFromJSON(raw); nil != err {
		return
	}
	if singleton.ChainMainName != object.doc.ChainID {
		err = ErrChainID
		return
	}
	appState := make(map[string]json.RawMessage)
	if err = json.Unmarshal(object.doc.AppState, &appState); nil != err {
		return
	}
	//transaction
	err = db.Transaction(func(tx *gorm.DB) (err error) {
		for _, method := range []func(cdc *amino.Codec,
			appState map[string]json.RawMessage,
			db *gorm.DB) (err error){
			object.processAccount,
			object.processAsset,
			object.processGenUtil,
		} {
			if err = method(cdc, appState, tx); nil != err {
				break
			}
		}
		if nil != err {
			// 创世完成
			err = service.NewSystem().UpdateLastBlockHeight(tx, 1)
		}
		return
	})
	return
}
