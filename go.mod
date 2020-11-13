module kds

go 1.15

replace github.com/KuChainNetwork/kuchain => github.com/KuChainNetwork/kratos v0.5.4

require (
	github.com/KuChainNetwork/kuchain v0.0.0-00010101000000-000000000000
	github.com/cosmos/cosmos-sdk v0.39.1
	github.com/gofiber/fiber/v2 v2.1.4
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/shopspring/decimal v1.2.0
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.33.8
	gorm.io/driver/mysql v1.0.2
	gorm.io/gorm v1.20.2
)
