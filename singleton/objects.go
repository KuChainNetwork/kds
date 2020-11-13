package singleton

import (
	"github.com/KuChainNetwork/kuchain/app"
	chainConfig "github.com/KuChainNetwork/kuchain/chain/config"
	constantsKeys "github.com/KuChainNetwork/kuchain/chain/constants/keys"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/tendermint/go-amino"
	"gorm.io/gorm"

	"kds/trieTree"

	"kds/db/model"
)

const (
	VersionName   = "kratos"
	ChainName     = "kts"
	ChainMainName = "kratos"
)

var (
	DB              *gorm.DB       // 数据库单例
	SystemModel     *model.System  // system数据模型单例
	NewDataNotifyCh chan struct{}  // 新数据通知通道
	Cdc             *amino.Codec   // 获取编解码器单例
	TXTrieTree      *trieTree.Node // 交易前缀树
	HeightTrieTree  *trieTree.Node // 高度前缀树
)

func init() {
	// 初始化编解码器
	version.Name = VersionName
	constantsKeys.ChainNameStr = ChainName
	constantsKeys.ChainMainNameStr = ChainMainName
	chainConfig.Bech32PrefixAccAddr = constantsKeys.ChainMainNameStr
	chainConfig.Bech32MainPrefix = constantsKeys.ChainMainNameStr
	chainConfig.Bech32PrefixAccAddr = chainConfig.Bech32MainPrefix
	chainConfig.Bech32PrefixAccPub = chainConfig.Bech32MainPrefix + chainConfig.PrefixPublic
	chainConfig.Bech32PrefixValAddr = chainConfig.Bech32MainPrefix + chainConfig.PrefixValidator + chainConfig.PrefixOperator
	chainConfig.Bech32PrefixValPub = chainConfig.Bech32MainPrefix + chainConfig.PrefixValidator + chainConfig.PrefixOperator + chainConfig.PrefixPublic
	chainConfig.Bech32PrefixConsAddr = chainConfig.Bech32MainPrefix + chainConfig.PrefixValidator + chainConfig.PrefixConsensus
	chainConfig.Bech32PrefixConsPub = chainConfig.Bech32MainPrefix + chainConfig.PrefixValidator + chainConfig.PrefixConsensus + chainConfig.PrefixPublic
	chainConfig.SealChainConfig()
	Cdc = app.MakeCodec()
	// 初始化新数据通知器
	NewDataNotifyCh = make(chan struct{}, 128)
	// 初始化交易前缀树
	TXTrieTree = trieTree.NewNode()
	// 初始化高度前缀树
	HeightTrieTree = trieTree.NewNode()
}
