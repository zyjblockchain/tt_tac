package logics

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
)

type SendTacTx struct {
	TacOrderId uint   `json:"tac_order_id" binding:"required,min=1"`
	Address    string `json:"address" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
	OrderType  int    `json:"order_type" binding:"required"`
}

func (s *SendTacTx) SendTacTx() (string, error) {
	// 查看地址是否存在数据库
	u, err := new(models.User).GetUserByAddress(s.Address)
	if err != nil {
		log.Errorf("地址不存在. error: %v", err)
		return "", err
	}
	// 验证password是否正确
	if !u.CheckPassword(s.Password) {
		log.Errorf("密码有误")
		return "", utils.VerifyPasswordErr
	}
	// 查看是那种跨链类型
	switch s.OrderType {
	case conf.EthToTtOrderType: // 从以太坊上的pala转移到tt链上的pala
		// 把以太坊上的pala发送到跨链转账的指定的中转地址
		chainNetUrl := conf.EthChainNet
		chainId := conf.EthChainID
		private, err := utils.DecryptPrivate(u.PrivateCrypted)
		if err != nil {
			log.Errorf("私钥aes解码失败， error: %v, address: %s", err, u.Address)
			return "", err
		}
		tokenAddress := common.HexToAddress(conf.EthPalaTokenAddress)
		toAddress := common.HexToAddress(conf.TacMiddleAddress)
		log.Infof("申请eth到tt的跨链转账用户开始发送eth_pala交易到中转地址, palaAmount: %s, address: %s", s.Amount, s.Address)
		txHash, err := s.send(chainNetUrl, int64(chainId), private, tokenAddress, toAddress)
		return txHash, err

	case conf.TtToEthOrderType: // 从tt链上的pala转移到eth上的pala
		// 把tt链上的pala发送到跨链转账的指定中转地址上
		chainNetUrl := conf.TTChainNet
		chainId := conf.TTChainID
		private, err := utils.DecryptPrivate(u.PrivateCrypted)
		if err != nil {
			log.Errorf("私钥aes解码失败， error: %v, address: %s", err, u.Address)
			return "", err
		}
		tokenAddress := common.HexToAddress(conf.TtPalaTokenAddress)
		toAddress := common.HexToAddress(conf.TacMiddleAddress)
		log.Infof("申请tt到eth的跨链转账用户开始发送tt_pala交易到中转地址, palaAmount: %s, address: %s", s.Amount, s.Address)
		txHash, err := s.send(chainNetUrl, int64(chainId), private, tokenAddress, toAddress)
		return txHash, err

	default:
		return "", errors.New("指定的order type不对")
	}
}

// send 发送交易封装
func (s *SendTacTx) send(chainNetUrl string, chainId int64, private string, tokenAddress common.Address, toAddress common.Address) (string, error) {
	client := transaction.NewChainClient(chainNetUrl, big.NewInt(chainId))
	defer client.Close()
	// 1. 判断地址是否有足够的amount
	palaBal, err := client.GetTokenBalance(common.HexToAddress(s.Address), tokenAddress)
	if err != nil {
		log.Errorf("获取pala balance error: %v, address: %s", err, s.Address)
		return "", err
	}
	amount, _ := new(big.Int).SetString(s.Amount, 10)
	if palaBal.Cmp(amount) < 0 {
		// pala余额不足
		log.Errorf("pala余额不足。address: %s, palaBalanc: %s, sendAmount: %s", s.Address, palaBal.String(), s.Amount)
		return "", errors.New("address上的pala余额不足")
	}
	// 2. 获取nonce
	nonce, err := client.GetNonce(common.HexToAddress(s.Address))
	if err != nil {
		log.Errorf("获取nonce失败, error: %v", err)
		return "", err
	}
	log.Infof("申请者跨链转账发送交易到中转地址的tx nonce: %d", nonce)
	// 3. 获取suggest gasPrice
	suggestPrice, err := client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Errorf("获取suggest gasPrice error: %v", err)
		return "", err
	}
	gasLimit := uint64(60000)
	gasPrice := suggestPrice.Mul(suggestPrice, big.NewInt(2)) // 两倍suggest gasPrice
	log.Infof("gas price: %v", gasPrice)
	// 发送交易
	signedTx, err := client.SendTokenTx(private, nonce, gasLimit, gasPrice, toAddress, tokenAddress, amount)
	if err != nil {
		log.Errorf("交易发送失败：error: %v", err)
		return "", err
	} else {
		return signedTx.Hash().String(), nil
	}
}
