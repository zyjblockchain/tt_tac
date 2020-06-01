package main

import (
	"context"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/routers"
	"github.com/zyjblockchain/tt_tac/utils"
	"github.com/zyjblockchain/tt_tac/utils/ding_robot"
	eth_watcher "github.com/zyjblockchain/tt_tac/utils/eth-watcher"
	"os"
	"strings"
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
			log.Errorf("以太坊上的block watcher error: %v", err)
			// 钉钉推送
			content := fmt.Sprintf("以太坊上的block watcher返回error,服务已经停止，请重启服务")
			_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, true)
			panic(err)
		}
	}()
	// 3.2 开启tt链上的监听
	ttChainApi := conf.TTChainNet
	ttChainWather := eth_watcher.NewHttpBasedEthWatcher(context.Background(), ttChainApi)
	go func() {
		err := ttChainWather.RunTillExit()
		if err != nil {
			log.Errorf("thundercore上的block watcher error: %v", err)
			// 钉钉推送
			content := fmt.Sprintf("thundercore上的block watcher返回error,服务已经停止，请重启服务")
			_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, true)
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

	// 5. 定时检查跨链转账和闪兑的中间地址的余额是否足够，如果不足，及时通知让其充值 todo 上线之后需要放开
	go func() {
		logics.CheckMiddleAddressBalance()
	}()

	// 6. 定时归集闪兑中的usdt到指定的地址中
	go func() {
		logics.DelayedCollectUsdtTx()
	}()

	// 对跨链转账的订单表中pending状态的订单处理
	logics.InitTacOrderState(ethToTtProcess, ttToEthProcess)
	// 对闪兑的订单表中pending状态的订单处理
	logics.InitFlashOrderState(flashChangeSrv)

	// 6. 启动gin服务
	routers.NewRouter(":3030")
}

// startBefore
func startBefore() {
	// 0. 初始化日志级别、格式、是否保存到文件
	log.Setup(log.LevelDebug, true, true)
	// 1. 读取配置文件
	if err := godotenv.Load("config"); err != nil {
		panic(err)
	}
	// 2. 从配置文件中加载数据
	initConf()
	// 3. 数据库链接
	// dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	models.InitDB(conf.Dsn)

	// 4. 从数据库中加载app的最新版本到内存中
	err := logics.InitAppVersionInfo()
	if err != nil {
		panic(fmt.Sprintf("从数据库中加载app最新版本信息失败，error: %v", err))
	}
}

func initConf() {
	var err error
	// 1. 跨链地址和私钥
	conf.TacMiddleAddress = os.Getenv("TacMiddleAddress")
	log.Infof("tac_middles_address: %s", conf.TacMiddleAddress)
	conf.TacMiddleAddressPrivate, err = utils.DecryptPrivate(os.Getenv("TacMiddleAddressPrivate"))
	if err != nil {
		panic(err)
	}
	// 检查私钥是否能还原出地址
	address, err := utils.PrivateToAddress(conf.TacMiddleAddressPrivate)
	if err != nil {
		panic(fmt.Sprintf("跨链转账中转地址私钥转地址失败，请检查跨链转账私钥是否正确。error: %v", err))
	}
	// 地址比较
	if strings.ToUpper(conf.TacMiddleAddress) != strings.ToUpper(address.String()) {
		panic(fmt.Sprintf("跨链转账中转地址私钥还原出的地址与配置的地址不一致, oldAddress: %s, newAddress: %s", conf.TacMiddleAddress, address.String()))
	}
	log.Infof("tac_middles_encryto_private: %s", conf.TacMiddleAddressPrivate)

	// 2. 数据库的dsn
	conf.Dsn = os.Getenv("MYSQL_DSN")

	// 3. 闪兑的地址和私钥
	conf.EthFlashChangeMiddleAddress = os.Getenv("EthFlashChangeMiddleAddress")
	log.Infof("flash_middles_address: %s", conf.EthFlashChangeMiddleAddress)
	conf.EthFlashChangeMiddlePrivate, err = utils.DecryptPrivate(os.Getenv("EthFlashChangeMiddlePrivate"))
	if err != nil {
		panic(err)
	}
	// 检查私钥是否能还原出地址
	address, err = utils.PrivateToAddress(conf.EthFlashChangeMiddlePrivate)
	if err != nil {
		panic(fmt.Sprintf("闪兑中转地址私钥转地址失败，请检查闪兑私钥是否正确。error: %v", err))
	}
	// 地址比较
	if strings.ToUpper(conf.EthFlashChangeMiddleAddress) != strings.ToUpper(address.String()) {
		panic(fmt.Sprintf("闪兑中转地址私钥还原出的地址与配置的地址不一致，oldAddress: %s, newAddress: %s", conf.EthFlashChangeMiddleAddress, address.String()))
	}
	log.Infof("flash_encryto_private: %s", conf.EthFlashChangeMiddlePrivate)

	// 4. 钉钉webHook
	conf.BalanceWebHook = os.Getenv("BalanceWebHook") // 中转地址余额不足的钉钉告警webHook
	log.Infof("BalanceWebHook: %s", conf.BalanceWebHook)
	conf.AbnormalWebHook = os.Getenv("AbnormalWebHook") // 其他异常的钉钉告警webHook
	log.Infof("AbnormalWebHook: %s", conf.AbnormalWebHook)

	// 5. usdt归集接收地址
	conf.ReceiveUSDTAddress = os.Getenv("ReceiveUSDTAddress") // usdt归集接收地址
	log.Infof("ReceiveUSDTAddress: %s", conf.ReceiveUSDTAddress)
}
