package utils

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/blockchain"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/plugin"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/structs"
	"math/big"
	"strings"
)

var (
	VerifyPasswordErr = errors.New("支付密码验证不通过")
)

var (
	VerifyParamsErrCode = 40001
	VerifyParamsErrMsg  = "参数校验失败"

	OrderLogicErrCode = 40002
	OrderLogicErrMsg  = "创建跨链转账订单失败"

	OrderFindErrCode = 40003
	OrderFindErrMsg  = "查询不到此订单"

	UserCreateErrCode = 40004
	UserCreateErrMsg  = "创建用户失败"

	UserLeadErrCode = 40005
	UserLeadErrMsg  = "导入用户失败"

	ExportPrivateErrCode = 40006
	ExportPrivateErrMsg  = "导出私钥失败"

	SendTacTxErrCode = 40007
	SendTacTxErrMsg  = "发送pala跨链交易失败"

	SendFlashChangeTxErrCode = 40008
	SendFlashChangeTxErrMsg  = "发送闪兑交易失败"

	GetBalanceErrCode = 40009
	GetBalanceErrMsg  = "get balance err"

	GetTokenBalanceErrCode = 40010
	GetTokenBalanceErrMsg  = "get token balance err"

	GetLatestPriceErrCode = 40011
	GetLatestPriceErrMsg  = "get latest price error"

	ExchangeOrderGetBatchErrCode = 40012
	ExchangeOrderGetBatchErrMsg  = "exchange order get batch error"

	TacOrderGetBatchErrCode = 40013
	TacOrderGetBatchErrMsg  = "tac order get batch error"

	ModifyPasswordErrCode = 40014
	ModifyPasswordErrMsg  = "modify password error"

	GetEthTokenTxRecordErrCode = 40015
	GetEthTokenTxRecordErrMsg  = "拉取 address pala token 收款记录 error"

	GetGasFeeErrCode = 40016
	GetGasFeeErrMsg  = "获取单笔交易的gas fee 失败"

	GetEthTxRecordErrCode = 40017
	GetEthTxRecordErrMsg  = "拉取 address eth 收款记录 error"

	EncryptoPrivErrCode = 40018
	EncryptoPrivErrMsg  = "对私钥进行对称加密失败"

	ModifyTacPalaServiceChargeErrCode = 40019
	ModifyTacPalaServiceChargeErrMsg  = "对私钥进行对称加密失败"

	ModifyPalaPriceComeUpRateErrCode = 40020
	ModifyPalaPriceComeUpRateErrMsg  = "ModifyPalaPriceComeUpRate error"

	CheckPasswordErrCode = 40021
	CheckPasswordErrMsg  = "check password error"

	SendPalaTransferErrCode = 40022
	SendPalaTransferErrMsg  = "send pala transfer failed"

	SendMainCoinTransferErrCode = 40023
	SendMainCoinTransferErrMsg  = "send main coin transfer failed"

	SendUsdtTransferErrCode = 40024
	SendUsdtTransferErrMsg  = "send pala transfer failed"

	SetAppVersionErrCode = 40025
	SetAppVersionErrMsg  = "设置app version失败"

	GetSendTransferRecordsErrCode = 40026
	GetSendTransferRecordsErrMsg  = "GetSendTransferRecords失败"
)

// FormatHex 去除前置的0
func FormatHex(s string) string {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	// 去除前置的所有0
	ss := strings.TrimLeft(s, "0")
	return "0x" + ss
}

func FormatAddressHex(s string) string {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	// 去除前置的所有0
	ss := strings.TrimLeft(s, "0")
	// 如果位数为基数则前置添加一个0
	if len(ss)%2 == 1 {
		ss = "0" + ss
	}
	return "0x" + ss
}

// PrivateToAddress 私钥转地址
func PrivateToAddress(private string) (common.Address, error) {
	p, err := crypto.ToECDSA(common.FromHex(private))
	if err != nil {
		return common.Address{}, err
	}
	addr := crypto.PubkeyToAddress(p.PublicKey)
	return addr, nil
}

// TransformAmount 跨链转账涉及到两条链的token兑换比例和gas fee的问题
func TransformAmount(oldAmount string, orderType int) string {
	// todo 目前不考虑兑换比例和交易gas fee的问题，后面有需求可以加上
	var fee *big.Int
	var newAmount string
	if orderType == conf.EthToTtOrderType { // 以太坊转入tt链
		fee = conf.EthToTtPalaCharge
	} else if orderType == conf.TtToEthOrderType { // tt 链转到以太坊
		fee = conf.TtToEthPalaCharge
	}
	amount, ok := new(big.Int).SetString(oldAmount, 10)
	if !ok {
		newAmount = oldAmount
	} else {
		// 判断amount 和 fee的大小
		if amount.Cmp(fee) <= 0 {
			newAmount = big.NewInt(0).String()
		} else {
			newAmount = new(big.Int).Sub(amount, fee).String()
		}

	}
	return newAmount
}

// UnitConversion 单位换算
// decimal 为换算token小数位数
// retainNum返回的最大截取小数位数
func UnitConversion(input string, decimal, retainNum int) string {
	inLen := len(input)
	// 前面添加0
	if inLen <= decimal {
		input = "0." + strings.Repeat("0", decimal-inLen) + input
	} else {
		input = input[0:inLen-decimal] + "." + input[inLen-decimal:]
	}

	if decimal > retainNum {
		// 截取小数位位数到保留位retainNum
		arr := strings.Split(input, ".")
		arr[1] = arr[1][:retainNum]
		if len(arr[1]) == 0 {
			input = arr[0]
		} else {
			input = arr[0] + "." + arr[1]
		}
	}
	// 如果input变成0.0000这种形式，只返回0.00
	ss := strings.Trim(input, "0")
	if ss == "." {
		return "0.00"
	}
	return input
}

// FormatTokenAmount
// input 单位换算之后的amount, "111.00111"
// decimal token的小数位数
func FormatTokenAmount(input string, decimal int) string {
	var result string
	arr := strings.Split(input, ".")
	if len(arr) == 1 {
		// 直接在后面添加0
		result = input + strings.Repeat("0", decimal)
	} else {
		if len(arr[1]) < decimal {
			arr[1] = arr[1] + strings.Repeat("0", decimal-len(arr[1]))
		} else if len(arr[1]) > decimal {
			// input的小数部分长度已经大于decimal则截取长度
			arr[1] = arr[1][:decimal]
		}
		result = arr[0] + arr[1]
	}
	// 去除前缀的0
	result = strings.TrimLeft(result, "0")
	return result
}

type TransferEvent struct {
	Token string
	From  string
	To    string
	Value decimal.Decimal
}

func ExtractERC20TransfersIfExist(r *structs.RemovableTxAndReceipt) (rst []TransferEvent) {
	transferEventSig := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	// todo a little weird
	if receipt, ok := r.Receipt.(*blockchain.EthereumTransactionReceipt); ok {
		for _, log := range receipt.Logs {
			if len(log.Topics) != 3 || log.Topics[0] != transferEventSig {
				continue
			}

			from := log.Topics[1]
			to := log.Topics[2]

			if amount, ok := plugin.HexToDecimal(log.Data); ok {
				rst = append(rst, TransferEvent{log.Address, from, to, amount})
			}
		}
	}

	return
}
