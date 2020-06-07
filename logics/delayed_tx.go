package logics

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/utils"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"time"
)

// 定时把闪兑的中间地址中的usdt发送到给定的地址中,private，和from为闪兑中间地址，to为usdt归集的接收地址
func SendAllUsdt(private, from, to string) (string, error) {
	client := transaction.NewChainClient(conf.EthChainNet, big.NewInt(int64(conf.EthChainID)))
	// 1. 检查from的usdt余额是否存在
	usdtBalance, err := client.GetTokenBalance(common.HexToAddress(from), common.HexToAddress(conf.EthUSDTTokenAddress))
	if err != nil {
		log.Errorf("获取eth usdt余额error: %v", err)
		return "", err
	}
	log.Infof("闪兑中转地址的usdt余额：%s", utils.UnitConversion(usdtBalance.String(), 6, 6))
	// 没有
	if usdtBalance.Cmp(big.NewInt(0)) <= 0 {
		log.Infof("闪兑中转地址的usdt余额为%s，则不需要进行归集", utils.UnitConversion(usdtBalance.String(), 6, 6))
		return "", nil
	}

	suggestPrice, err := client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Errorf("获取suggest gasPrice error: %v", err)
		return "", err
	}
	gasLimit := uint64(70000)
	gasPrice := suggestPrice.Mul(suggestPrice, big.NewInt(3)) // suggest gasPrice
	// 3.3 获取nonce
	nonce, err := client.GetLatestNonce(from)
	if err != nil {
		log.Errorf("usdt归集获取nonce失败, error: %v,address: %s", err, from)
		return "", err
	}
	log.Infof("开始发送闪兑中间地址usdt转账归集交易")
	tx, err := client.SendTokenTx(private, nonce, gasLimit, gasPrice, common.HexToAddress(to), common.HexToAddress(conf.EthUSDTTokenAddress), usdtBalance)
	if err != nil {
		// 回归nonce
		client.SetFailNonce(from, nonce)
		return "", err
	}
	log.Infof("txHash: %s", tx.Hash().String())
	// 监听交易链上状态
	count := 0
	for {
		if count > 10 {
			// 超时
			// 回归nonce
			client.SetFailNonce(from, nonce)
			return "", errors.New("从闪兑中间地址归集usdt到制定地址交易发送超时")
		}

		time.Sleep(15 * time.Second)
		_, isPending, err := client.Client.TransactionByHash(context.Background(), tx.Hash())
		if err == nil && !isPending {
			// 查询到了交易，修改交易状态为成功
			log.Infof("从闪兑中间地址归集usdt到制定地址交易成功,txHash: %s", tx.Hash().String())
			return tx.Hash().String(), nil
		}
		// 增加count
		count++
	}
}

// 把闪兑的中间地址中的usdt定时归集到指定的接收地址中
func DelayedCollectUsdtTx() {
	private := conf.EthFlashChangeMiddlePrivate
	from := conf.EthFlashChangeMiddleAddress
	to := conf.ReceiveUSDTAddress
	// 每一分钟归集一次
	for {
		_, err := SendAllUsdt(private, from, to)
		if err != nil {
			log.Errorf("发送usdt归集交易失败；error: %v", err)
		}
		time.Sleep(300 * time.Second)
	}

}
