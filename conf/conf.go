package conf

import (
	"math/big"
)

const (
	TTChainNet = "https://mainnet-rpc.thundercore.com"
	TTChainID  = 108
	TTChainTag = 77

	// // 测试网
	// EthChainNet = "https://rinkeby.infura.io/v3/36b98a13557c4b8583d57934ede2f74d"
	// EthChainID  = 4

	// 备用节点
	// EthChainNet = "https://mainnet.infura.io/v3/36b98a13557c4b8583d57934ede2f74d" // 18382255942

	// 主网
	EthChainNet = "https://mainnet.infura.io/v3/7bbf73a8855d4c0491f93e6dc498360d" // 1263344073
	EthChainID  = 1                                                               // 以太坊的主网chainId == 1

	EthChainTag = 17
)

// 正式环境的token
const (
	EthPalaTokenAddress = "0xD20fb5cf926Dc29c88f64725e6f911f40f7bf531" // 以太坊主网上的pala合约地址
	TtPalaTokenAddress  = "0xeff6f1612d03205BA5E8d26cAc1397bf778ab1AC" // tt主网上的pala合约地址
	EthUSDTTokenAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7" // 以太坊主网上的USDT合约地址
)

// // 测试环境的token
// const (
// 	EthPalaTokenAddress = "0x03332638A6b4F5442E85d6e6aDF929Cd678914f1" // 以太坊上的pala erc20 地址，目前是测试环境 以太坊rinkeby 上的test3
// 	TtPalaTokenAddress  = "0x087cC4Aaa83aCA54bDCC89920483c8e2a30Bc47c" // tt上的pala 地址，目前是测试环境tt主网上的sandy
// 	EthUSDTTokenAddress = "0xD1Df5b185198F3c6Da74e93B36b7E29523c265F0" // 以太坊上的usdt erc20 地址, 目前是以太坊测试网的测试token
// )

const (
	EthToTtOrderType = 1 // 以太坊转tt链
	TtToEthOrderType = 2 // tt链转以太坊
)

// 跨链转账扣除pala手续费数量
var (
	EthToTtPalaCharge          = big.NewInt(1 * 10000000) // 以太坊转tt链中间扣除的pala默认手续费，有接口可以随时修改
	TtToEthPalaCharge          = big.NewInt(1 * 10000000) // tt链转以太坊中间扣除的pala默认手续费，有接口可以随时修改
	FlashPalaToUsdtPriceChange = float64(1.01)            // 在闪兑中pala的价格需要增大来展示给用户，通过这种方式变相收取闪兑的手续费。默认上浮1%，有接口可以随时修改
)

// 需要配置文件读取
var (
	// 跨链转账中转地址，tt链和以太坊链共用一个地址，方便管理
	TacMiddleAddress        = ""
	TacMiddleAddressPrivate = ""
	Dsn                     = ""

	// eth usdt -> pala闪兑中转地址
	EthFlashChangeMiddleAddress = ""
	EthFlashChangeMiddlePrivate = ""
)
var BalanceWebHook = ""     // 中转地址余额不足的钉钉告警webHook
var AbnormalWebHook = ""    // 其他异常的钉钉告警webHook
var ReceiveUSDTAddress = "" // usdt归集接收地址
