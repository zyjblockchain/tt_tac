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
