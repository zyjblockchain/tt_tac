package logics

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
)

// 发送pala转账交易
type PalaTransfer struct {
	FromAddress string `json:"from_address" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
}

func (p *PalaTransfer) SendPalaTx(chainTag int) (string, error) {
	// 0. 验证支付密码
	user, err := new(models.User).GetUserByAddress(p.FromAddress)
	if err != nil {
		log.Errorf("通过address从表中查询user失败， err: %v, address: %s", err, p.FromAddress)
		return "", err
	}
	if !user.CheckPassword(p.Password) {
		log.Errorf("密码有误")
		return "", utils.VerifyPasswordErr
	}

	// new client
	var client *transaction.ChainClient
	if chainTag == conf.EthChainTag {
		client = transaction.NewChainClient(conf.EthChainNet, big.NewInt(int64(conf.EthChainID)))
	} else if chainTag == conf.TTChainTag {
		client = transaction.NewChainClient(conf.TTChainNet, big.NewInt(int64(conf.TTChainID)))
	} else {
		return "", errors.New("不存在的chainTag")
	}

	// 1. 检查from的pala余额是否足够
	palaBalance, err := client.GetTokenBalance(common.HexToAddress(p.FromAddress), common.HexToAddress(conf.EthPalaTokenAddress))
	if err != nil {
		log.Errorf("获取eth pala余额error: %v", err)
		return "", err
	}
	// 2. 比较pala余额
	amount, _ := new(big.Int).SetString(p.Amount, 10)
	if palaBalance.Cmp(amount) < 0 {
		log.Errorf("eth pala转账余额不足；转账amount: %s, pala余额：%s, address: %s", p.Amount, palaBalance.String(), p.FromAddress)
		return "", errors.New(fmt.Sprintf("eth pala转账余额不足；转账amount: %s, pala余额：%s, address: %s", p.Amount, palaBalance.String(), p.FromAddress))
	}

	// 3. 发送交易
	// 3.1 解码私钥
	private, err := utils.DecryptPrivate(user.PrivateCrypted)
	if err != nil {
		log.Errorf("私钥aes解码失败， error: %v, address: %s", err, user.Address)
		return "", err
	}
	// 3.2 获取nonce
	nonce, err := client.GetNonce(common.HexToAddress(p.FromAddress))
	if err != nil {
		log.Errorf("获取nonce失败, error: %v,address: %s", err, p.FromAddress)
		return "", err
	}
	// 3.3 获取suggest gasPrice
	suggestPrice, err := client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Errorf("获取suggest gasPrice error: %v", err)
		return "", err
	}
	gasLimit := uint64(60000)
	gasPrice := suggestPrice.Mul(suggestPrice, big.NewInt(2)) // 两倍suggest gasPrice
	tx, err := client.SendTokenTx(private, nonce, gasLimit, gasPrice, common.HexToAddress(p.ToAddress), common.HexToAddress(conf.EthPalaTokenAddress), amount)
	if err != nil {
		log.Errorf("发送eth pala交易失败；error: %v", err)
	}
	return tx.Hash().String(), nil
}
