package main

import (
	"context"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/routers"
	eth_watcher "github.com/zyjblockchain/tt_tac/utils/eth-watcher"
)

func main() {
	// 1. 初始化日志级别、格式、是否保存到文件
	log.Setup(log.LevelDebug, true, true)
	// 2. 数据库链接
	// dsn := "tac_user:tac_good123@tcp(rm-j6c49n23e4d07l8ijmo.mysql.rds.aliyuncs.com:3306)/tac_db?charset=utf8&parseTime=True&loc=Local"
	dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	models.InitDB(dsn)

	// 3. 启动跨链转账服务
	// ethChainApi := "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4"
	ethChainApi := conf.EthChainNet
	ttChainApi := conf.TTChainNet
	ethChainWather := eth_watcher.NewHttpBasedEthWatcher(context.Background(), ethChainApi)
	go func() {
		err := ethChainWather.RunTillExit()
		if err != nil {
			panic(err)
		}
	}()
	ttChainWather := eth_watcher.NewHttpBasedEthWatcher(context.Background(), ttChainApi)
	go func() {
		err := ttChainWather.RunTillExit()
		if err != nil {
			panic(err)
		}
	}()

	ethPalaTokenAddress := "0x03332638A6b4F5442E85d6e6aDF929Cd678914f1" // 测试环境 以太坊rinkeby 上的test3
	ttPalaTokenAddress := "0x087cC4Aaa83aCA54bDCC89920483c8e2a30Bc47c"  // 测试环境tt主网上的sandy

	// new eth -> tt process
	ethToTtProcess := logics.NewTacProcess(ethChainApi, ethPalaTokenAddress, ttPalaTokenAddress, conf.MiddleAddress, ethChainWather, ttChainWather)
	ethToTtProcess.ListenErc20CollectionAddress()
	// new tt -> eth process
	ttToEthProcess := logics.NewTacProcess(ttChainApi, ttPalaTokenAddress, ethPalaTokenAddress, conf.MiddleAddress, ttChainWather, ethChainWather)
	ttToEthProcess.ListenErc20CollectionAddress()

	// 4. 启动服务
	routers.NewRouter(":3000")
}
