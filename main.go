package main

import (
	"context"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/routers"
	"github.com/zyjblockchain/tt_tac/utils"
	eth_watcher "github.com/zyjblockchain/tt_tac/utils/eth-watcher"
	"os"
)

func main() {
	// 1. 服务开始之前的init
	startBefore()

	// 3. 启动跨链转账服务
	// ethChainApi := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	// 3.1 开启eth链的监听
	ethChainApi := conf.EthChainNet
	ethChainWather := eth_watcher.NewHttpBasedEthWatcher(context.Background(), ethChainApi)
	go func() {
		err := ethChainWather.RunTillExit()
		if err != nil {
			panic(err)
		}
	}()
	// 3.2 开启tt链上的监听
	ttChainApi := conf.TTChainNet
	ttChainWather := eth_watcher.NewHttpBasedEthWatcher(context.Background(), ttChainApi)
	go func() {
		err := ttChainWather.RunTillExit()
		if err != nil {
			panic(err)
		}
	}()

	// 3.3 启动 eth -> tt 跨链process
	ethToTtProcess := logics.NewTacProcess(ethChainApi, conf.EthPalaTokenAddress, conf.TtPalaTokenAddress, conf.TacMiddleAddress, ethChainWather, ttChainWather)
	ethToTtProcess.ListenErc20CollectionAddress()
	// 3.3 启动 tt -> eth 跨链process
	ttToEthProcess := logics.NewTacProcess(ttChainApi, conf.TtPalaTokenAddress, conf.EthPalaTokenAddress, conf.TacMiddleAddress, ttChainWather, ethChainWather)
	ttToEthProcess.ListenErc20CollectionAddress()

	// 4. 启动闪兑服务
	flashChangeSrv := logics.NewWatchFlashChange(conf.EthUSDTTokenAddress, conf.EthPalaTokenAddress, ethChainWather)
	flashChangeSrv.ListenFlashChangeTx()

	// 5. 定时检查跨链转账和闪兑的中间地址的余额是否足够，如果不足，及时通知让其充值
	go func() {
		// logics.CheckMiddleAddressBalance()
	}()

	// 6. 启动gin服务
	routers.NewRouter(":3030")
}

// startBefore
func startBefore() {
	// 0. 初始化日志级别、格式、是否保存到文件
	log.Setup(log.LevelDebug, true, true)
	// 1. 读取配置文件
	if err := godotenv.Load("config_dev"); err != nil {
		panic(err)
	}
	// 2. 从配置文件中加载数据
	initConf()
	// 2. 数据库链接
	// dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	models.InitDB(conf.Dsn)
}

func initConf() {
	var err error
	conf.TacMiddleAddressPrivate = os.Getenv("TacMiddleAddressPrivate")
	conf.TacMiddleAddress = os.Getenv("TacMiddleAddress")
	// todo 正式环境配置文件中的private是aes加密之后的，所以这里需要解密
	conf.TacMiddleAddressPrivate, err = utils.DecryptPrivate(os.Getenv("TacMiddleAddressPrivate"))
	if err != nil {
		panic(err)
	}

	conf.Dsn = os.Getenv("MYSQL_DSN")

	conf.EthFlashChangeMiddleAddress = os.Getenv("EthFlashChangeMiddleAddress")
	conf.EthFlashChangeMiddlePrivate = os.Getenv("EthFlashChangeMiddlePrivate")
	conf.EthFlashChangeMiddlePrivate, err = utils.DecryptPrivate(os.Getenv("EthFlashChangeMiddlePrivate"))
	if err != nil {
		panic(err)
	}
}
