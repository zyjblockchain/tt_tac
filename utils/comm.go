package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zyjblockchain/tt_tac/conf"
	"math/big"
	"strings"
)

var WebHook = "https://oapi.dingtalk.com/robot/send?access_token=b4ff4c39e202803e650886c6a93003e5423796525d9ff1f777c13a2a03762da8"

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
)

// FormatHex 去除前置的0
func FormatHex(s string) string {
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
		fee = big.NewInt(1 * 100000000) // 1 pala
	} else if orderType == conf.TtToEthOrderType { // tt 链转到以太坊
		fee = big.NewInt(3 * 100000000) // 3 pala
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
