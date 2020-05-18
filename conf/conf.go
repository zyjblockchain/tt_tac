package conf

import "os"

const (
	TTChainNet = "https://mainnet-rpc.thundercore.com"
	TTChainID  = 108
	TTChainTag = 11

	EthChainNet = "https://rinkeby.infura.io/v3/36b98a13557c4b8583d57934ede2f74d"
	EthChainID  = 4
	EthChainTag = 22
)

const (
	EthPalaTokenAddress = "0x03332638A6b4F5442E85d6e6aDF929Cd678914f1" // 测试环境 以太坊rinkeby 上的test3
	TtPalaTokenAddress  = "0x087cC4Aaa83aCA54bDCC89920483c8e2a30Bc47c" // 测试环境tt主网上的sandy
)

const (
	EthToTtOrderType = 1 // 以太坊转tt链
	TtToEthOrderType = 2 // tt链转以太坊
)

var (
	MiddleAddress        = ""
	MiddleAddressPrivate = ""
	Dsn                  = ""
)

func InitConf() {
	MiddleAddress = os.Getenv("MiddleAddress")
	MiddleAddressPrivate = os.Getenv("MiddleAddressPrivate")
	Dsn = os.Getenv("MYSQL_DSN")
}
