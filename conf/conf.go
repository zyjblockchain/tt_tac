package conf

import "os"

const (
	TTChainNet = "https://mainnet-rpc.thundercore.com"
	TTChainID  = 108
	TTChainTag = 77

	EthChainNet = "https://rinkeby.infura.io/v3/36b98a13557c4b8583d57934ede2f74d"
	EthChainID  = 4
	EthChainTag = 17
)

const (
	EthPalaTokenAddress = "0x03332638A6b4F5442E85d6e6aDF929Cd678914f1" // 以太坊上的pala erc20 地址，目前是测试环境 以太坊rinkeby 上的test3
	TtPalaTokenAddress  = "0x087cC4Aaa83aCA54bDCC89920483c8e2a30Bc47c" // tt上的pala 地址，目前是测试环境tt主网上的sandy
	EthUSDTTokenAddress = "0xD1Df5b185198F3c6Da74e93B36b7E29523c265F0" // 以太坊上的usdt erc20 地址, 目前是以太坊测试网的测试token
)

const (
	EthToTtOrderType = 1 // 以太坊转tt链
	TtToEthOrderType = 2 // tt链转以太坊
)

var (
	// 跨链转账中转地址，tt链和以太坊链共用一个地址，方便管理
	TacMiddleAddress        = ""
	TacMiddleAddressPrivate = ""
	Dsn                     = ""
)

var (
	// eth usdt -> pala闪兑中转地址
	EthFlashChangeMiddleAddress = ""
	EthFlashChangeMiddlePrivate = ""
)

func InitConf() {
	TacMiddleAddress = os.Getenv("TacMiddleAddress")
	TacMiddleAddressPrivate = os.Getenv("TacMiddleAddressPrivate")
	Dsn = os.Getenv("MYSQL_DSN")
	EthFlashChangeMiddleAddress = os.Getenv("EthFlashChangeMiddleAddress")
	EthFlashChangeMiddlePrivate = os.Getenv("EthFlashChangeMiddlePrivate")
}
