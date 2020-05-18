package logics

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
)

// 闪兑 usdt -> eth_pala
type FlashChange struct {
	FromTokenAddress string `json:"from_token_address"`                   // usdt token address，目前是写死，后面为了扩展会要求前端传
	ToTokenAddress   string `json:"to_token_address"`                     // eth_pala token address，目前是写死，后面为了扩展会要求前端传
	OperateAddress   string `json:"operate_address" binding:"required"`   // 操作地址
	FromTokenAmount  string `json:"from_token_amount" binding:"required"` // 需要兑换数量,usdt的数量
	ToTokenAmount    string `json:"to_token_amount"`                      // 转换之后的数量, usdt按照实时价格转换pala的数量。todo 为收取手续费，所以实时价格需要提高1%进行兑换
}

func (f *FlashChange) FlashChange() error {
	// 1. 查看operateAddress 是否存在正在进行中的闪兑订单
	_, err := (&models.FlashChangeOrder{
		OperateAddress:   f.OperateAddress,
		FromTokenAddress: f.FromTokenAddress,
		ToTokenAddress:   f.ToTokenAddress,
		State:            1,
	}).Get()
	if err == nil {
		// 存在则返回
		log.Errorf("存在正在进行中的同类型的闪兑订单。operateAddress: %s, fromToken: %s, toToken: %s", f.OperateAddress, f.FromTokenAddress, f.ToTokenAddress)
		return errors.New(fmt.Sprintf("存在正在进行中的同类型的闪兑订单。operateAddress: %s, fromToken: %s, toToken: %s", f.OperateAddress, f.FromTokenAddress, f.ToTokenAddress))
	}

	// 2. 查看operateAddress是否有足够多的FromTokenAmount
	client := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer client.Close()
	fromTokenBalance, err := client.GetTokenBalance(common.HexToAddress(f.OperateAddress), common.HexToAddress(f.FromTokenAddress))
	if err != nil {
		log.Errorf("get token balance error: %v, address: %s, tokenAddress: %s", err, f.OperateAddress, f.FromTokenAddress)
		return err
	}
	fromTokenAmount, _ := new(big.Int).SetString(f.FromTokenAmount, 10)
	if fromTokenBalance.Cmp(fromTokenAmount) < 0 {
		// 余额不足
		log.Errorf("账户中闪兑的from token 余额不足。address: %s, tokenAddress: %s, fromTokenBalance：%s, fromTokenAmount: %s",
			f.OperateAddress, f.FromTokenAddress, fromTokenBalance.String(), f.FromTokenAmount)
		return errors.New(fmt.Sprintf("账户中闪兑的from token 余额不足。address: %s, tokenAddress: %s, fromTokenBalance：%s, fromTokenAmount: %s",
			f.OperateAddress, f.FromTokenAddress, fromTokenBalance.String(), f.FromTokenAmount))
	}

	// 3. 执行闪兑process
	// 3.1 保存闪兑订单
	fo := &models.FlashChangeOrder{
		OperateAddress:   f.OperateAddress,
		FromTokenAddress: f.FromTokenAddress,
		ToTokenAddress:   f.ToTokenAddress,
		FromTokenAmount:  f.FromTokenAmount,
		ToTokenAmount:    f.ToTokenAmount,
		State:            1,
	}
	if err := fo.Create(); err != nil {
		log.Errorf("保存flash change order error: %v", err)
		return err
	}
	// 3.2 发送交易

}

// 开启一个wather来监听闪兑的交易
