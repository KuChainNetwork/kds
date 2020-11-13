package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"

	"kds/blockAnalyser"
	"kds/blockDataGetter"
	"kds/db"
	"kds/genesis"
	"kds/singleton"
	"kds/txAnalyser"
	"kds/web"
)

var (
	username   = "dev"                         // 数据库用户名
	password   = "dev"                         // 数据库密码
	host       = "127.0.0.1"                   // 数据库主机
	database   = "dev"                         // 数据库名
	port       = 3307                          // 数据库端口
	retryTimes = 60                            //
	batchLimit = int64(1024)                   // 数据批处理大小限制
	chainID    = "kratos"                      // 链ID
	nodeURI    = "http://121.89.211.107:34568" // 节点URI
	maxGetters = 256                           // 最大区块获取并发数
	httpPort   = 8083                          // API服务端口号
)

// OnExit 退出
func OnExit(fn func()) {
	sigCh := make(chan os.Signal, 32)
	signal.Notify(sigCh)
sigLoop:
	for sig := range sigCh {
		// urgent I/O condition
		if syscall.Signal(0x17) == sig {
			continue
		}
		glog.Errorln("signal:", sig)
		switch sig {
		case os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT:
			signal.Stop(sigCh)
			break sigLoop
		}
	}
	close(sigCh)
	fn()
}

func main() {
	flag.StringVar(&username, "username", "dev", "mysql database username")
	flag.StringVar(&password, "password", "dev", "mysql database password")
	flag.StringVar(&host, "host", "localhost", "mysql database host")
	flag.IntVar(&port, "port", 3306, "mysql database port")
	flag.IntVar(&retryTimes, "retryTimes", 60, "connect mysql retry times")
	flag.Int64Var(&batchLimit, "batchLimit", 1024, "analyzer batch process count limit")
	flag.StringVar(&chainID, "chainId", "kratos", "chain id")
	flag.StringVar(&nodeURI, "nodeUri", "http://127.0.0.1:26657", "node uri")
	flag.IntVar(&maxGetters, "maxGetters", 256, "pull block and block result concurrency number")
	flag.IntVar(&httpPort, "httpPort", 8080, "restful service port")
	flag.Parse() // 解析命令行
	// 初始化数据库
	err := db.Initialize(username, password, host, database, port, retryTimes)
	if nil != err {
		glog.Fatalln(err)
		return
	}
	defer db.Dispose()
	// 创世
	var done bool
	if err, done = genesis.New(fmt.Sprintf("%s/genesis", nodeURI)).
		Initialize(singleton.DB, singleton.Cdc, singleton.SystemModel.LastBlockHeight); nil != err {
		glog.Fatalln(err)
		return
	} else if done {
		singleton.SystemModel.LastBlockHeight = 1
	}
	// 开始分析
	blockAnalyserObject := blockAnalyser.New(singleton.DB, singleton.Cdc, singleton.NewDataNotifyCh)
	if err = blockAnalyserObject.Start(batchLimit); nil != err {
		glog.Fatalln(err)
		return
	}
	txAnalyserObject := txAnalyser.New(singleton.DB, singleton.Cdc, singleton.NewDataNotifyCh)
	if err = txAnalyserObject.Start(batchLimit); nil != err {
		glog.Fatalln(err)
		return
	}
	// 拉取区块数据
	gatterGroup := blockDataGetter.NewGetterGroup(chainID, nodeURI, singleton.Cdc, singleton.DB, maxGetters)
	if err = gatterGroup.Start(singleton.SystemModel.LastBlockHeight, singleton.NewDataNotifyCh); nil != err {
		glog.Fatalln(err)
		return
	}
	// HTTP服务
	httpService := web.NewHTTPServer(httpPort, singleton.DB)
	if err = httpService.Start(); nil != err {
		glog.Fatalln(err)
		return
	}
	// 等待退出
	OnExit(func() {
		if err = httpService.Stop(); nil != err {
			glog.Errorln(err)
		}
		if err = gatterGroup.Stop(); nil != err {
			glog.Errorln(err)
		}
		close(singleton.NewDataNotifyCh)
		if err = blockAnalyserObject.Stop(); nil != err {
			glog.Errorln(err)
		}
		if err = txAnalyserObject.Stop(); nil != err {
			glog.Errorln(err)
		}
	})
}
