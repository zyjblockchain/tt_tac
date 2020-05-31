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
	"time"
)

// 发送pala转账交易
type PalaTransfer struct {
	FromAddress string `json:"from_address" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
}

// SendPalaTx
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
	var palaTokenAddress string
	var client *transaction.ChainClient
	if chainTag == conf.EthChainTag {
		client = transaction.NewChainClient(conf.EthChainNet, big.NewInt(int64(conf.EthChainID)))
		palaTokenAddress = conf.EthPalaTokenAddress
	} else if chainTag == conf.TTChainTag {
		client = transaction.NewChainClient(conf.TTChainNet, big.NewInt(int64(conf.TTChainID)))
		palaTokenAddress = conf.TtPalaTokenAddress
	} else {
		return "", errors.New("不存在的chainTag")
	}

	// 1. 检查from的pala余额是否足够
	palaBalance, err := client.GetTokenBalance(common.HexToAddress(p.FromAddress), common.HexToAddress(palaTokenAddress))
	if err != nil {
		log.Errorf("获取pala余额error: %v", err)
		return "", err
	}
	// 2. 比较pala余额
	amount, _ := new(big.Int).SetString(p.Amount, 10)
	if palaBalance.Cmp(amount) < 0 {
		log.Errorf(" pala转账余额不足；转账amount: %s, pala余额：%s, address: %s", p.Amount, palaBalance.String(), p.FromAddress)
		return "", errors.New(fmt.Sprintf("pala转账余额不足；转账amount: %s, pala余额：%s, address: %s", p.Amount, palaBalance.String(), p.FromAddress))
	}

	// 3. 发送交易
	// 3.1 解码私钥
	private, err := utils.DecryptPrivate(user.PrivateCrypted)
	if err != nil {
		log.Errorf("私钥aes解码失败， error: %v, address: %s", err, user.Address)
		return "", err
	}

	// 3.2 获取suggest gasPrice
	suggestPrice, err := client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Errorf("获取suggest gasPrice error: %v", err)
		return "", err
	}
	gasLimit := uint64(60000)
	gasPrice := suggestPrice.Mul(suggestPrice, big.NewInt(2)) // 两倍suggest gasPrice

	// 3.3 获取nonce
	nonce, err := client.GetLatestNonce(p.FromAddress)
	if err != nil {
		log.Errorf("获取nonce失败, error: %v,address: %s", err, p.FromAddress)
		return "", err
	}
	// 3.4 把交易info保存到sendTransfer表
	ss := &models.SendTransfer{
		FromAddress:  p.FromAddress,
		ToAddress:    p.ToAddress,
		Amount:       p.Amount,
		TokenAddress: palaTokenAddress,
		TxHash:       "",
		OwnChain:     chainTag,
		CoinType:     2, // pala币
		TxStatus:     0,
		ErrMsg:       "",
	}
	err = ss.Create()
	if err != nil {
		client.SetFailNonce(p.FromAddress, nonce)
		log.Errorf("创建sendTransfer记录失败； error: %v", err)
		return "", err
	}
	// 4. 发送交易
	tx, err := client.SendTokenTx(private, nonce, gasLimit, gasPrice, common.HexToAddress(p.ToAddress), common.HexToAddress(palaTokenAddress), amount)
	if err != nil {
		// 交易发送失败，设置交易记录状态
		_ = ss.Update(models.SendTransfer{TxStatus: 2, ErrMsg: err.Error()})
		// 回归nonce
		client.SetFailNonce(p.FromAddress, nonce)
		log.Errorf("发送eth pala交易失败；error: %v", err)
		return "", err
	}
	// 4.1 更新交易hash
	_ = ss.Update(models.SendTransfer{TxHash: tx.Hash().String()})
	// 4.2 保存交易到kv表中
	byteTx, _ := tx.MarshalJSON()
	_ = models.SetKv(tx.Hash().String(), byteTx)

	// 5. 监听交易链上状态
	go func() {
		count := 0
		for {
			if count > 10 {
				// 超时
				_ = ss.Update(models.SendTransfer{TxStatus: 3})
				// 回归nonce
				client.SetFailNonce(p.FromAddress, nonce)
				return
			}

			time.Sleep(15 * time.Second)
			_, isPending, err := client.Client.TransactionByHash(context.Background(), tx.Hash())
			if err == nil && !isPending {
				// 查询到了交易，修改交易状态为成功
				log.Infof("链上查询到了转账交易；txHash: %s", tx.Hash().String())
				_ = ss.Update(models.SendTransfer{TxStatus: 1})
				return
			}
			// 增加count
			count++
		}
	}()

	return tx.Hash().String(), nil
}

type CoinTransfer struct {
	FromAddress string `json:"from_address" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
}

// SendMainNetCoin 发送tt币或者是eth
func (c *CoinTransfer) SendMainNetCoinTransfer(chainTag int) (string, error) {
	// 0. 验证支付密码
	user, err := new(models.User).GetUserByAddress(c.FromAddress)
	if err != nil {
		log.Errorf("通过address从表中查询user失败， err: %v, address: %s", err, c.FromAddress)
		return "", err
	}
	if !user.CheckPassword(c.Password) {
		log.Errorf("密码有误")
		return "", utils.VerifyPasswordErr
	}

	// 1. new client
	var client *transaction.ChainClient
	if chainTag == conf.EthChainTag {
		client = transaction.NewChainClient(conf.EthChainNet, big.NewInt(int64(conf.EthChainID)))
	} else if chainTag == conf.TTChainTag {
		client = transaction.NewChainClient(conf.TTChainNet, big.NewInt(int64(conf.TTChainID)))
	} else {
		return "", errors.New("不存在的chainTag")
	}

	// getBalance
	balance, err := client.Client.BalanceAt(context.Background(), common.HexToAddress(c.FromAddress), nil)
	if err != nil {
		log.Errorf("获取主网币余额error: %v", err)
		return "", err
	}

	// 2. 查看余额是否足够
	amount, _ := new(big.Int).SetString(c.Amount, 10)
	if balance.Cmp(amount) <= 0 {
		log.Errorf("主网币 转账余额不足；转账amount: %s, 余额：%s, address: %s", c.Amount, balance.String(), c.FromAddress)
		return "", errors.New(fmt.Sprintf("主网币 转账余额不足；转账amount: %s, 余额：%s, address: %s", c.Amount, balance.String(), c.FromAddress))
	}

	// 3. 发送转账交易
	// 3.1 解码私钥
	private, err := utils.DecryptPrivate(user.PrivateCrypted)
	if err != nil {
		log.Errorf("私钥aes解码失败， error: %v, address: %s", err, user.Address)
		return "", err
	}

	// 3.2 获取suggest gasPrice
	suggestPrice, err := client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Errorf("获取suggest gasPrice error: %v", err)
		return "", err
	}
	gasLimit := uint64(22000)
	gasPrice := suggestPrice.Mul(suggestPrice, big.NewInt(2)) // 两倍suggest gasPrice

	// 3.3 获取nonce
	nonce, err := client.GetLatestNonce(c.FromAddress)
	if err != nil {
		log.Errorf("获取nonce失败, error: %v,address: %s", err, c.FromAddress)
		return "", err
	}

	// 3.4 把交易info保存到sendTransfer
	ss := &models.SendTransfer{
		FromAddress:  c.FromAddress,
		ToAddress:    c.ToAddress,
		Amount:       c.Amount,
		TokenAddress: "",
		TxHash:       "",
		OwnChain:     chainTag,
		CoinType:     1,
		TxStatus:     0,
		ErrMsg:       "",
	}
	err = ss.Create()
	if err != nil {
		client.SetFailNonce(c.FromAddress, nonce)
		log.Errorf("创建sendTransfer记录失败； error: %v", err)
		return "", err
	}

	// 4. 发送交易
	tx, err := client.SendNormalTx(private, nonce, gasLimit, gasPrice, common.HexToAddress(c.ToAddress), amount)
	if err != nil {
		// 交易发送失败，设置交易记录状态
		_ = ss.Update(models.SendTransfer{TxStatus: 2, ErrMsg: err.Error()})
		// 回归nonce
		client.SetFailNonce(c.FromAddress, nonce)
		log.Errorf("发送主网币交易失败；error: %v", err)
		return "", err
	}
	// 4.1 更新交易hash
	_ = ss.Update(models.SendTransfer{TxHash: tx.Hash().String()})
	// 4.2 保存交易到kv表中
	byteTx, _ := tx.MarshalJSON()
	_ = models.SetKv(tx.Hash().String(), byteTx)
	// 5. 监听交易链上状态
	go func() {
		count := 0
		for {
			if count > 10 {
				// 超时
				_ = ss.Update(models.SendTransfer{TxStatus: 3})
				// 回归nonce
				client.SetFailNonce(c.FromAddress, nonce)
				return
			}

			time.Sleep(15 * time.Second)
			_, isPending, err := client.Client.TransactionByHash(context.Background(), tx.Hash())
			if err == nil && !isPending {
				// 查询到了交易，修改交易状态为成功
				_ = ss.Update(models.SendTransfer{TxStatus: 1})
				log.Infof("链上查询到了转账交易；txHash: %s", tx.Hash().String())
				return
			}
			// 增加count
			count++
		}
	}()

	return tx.Hash().String(), nil
}

type EthUsdtTransfer struct {
	FromAddress string `json:"from_address" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
}

func (c *EthUsdtTransfer) SendEthUsdtTransfer() (string, error) {
	// 0. 验证支付密码
	user, err := new(models.User).GetUserByAddress(c.FromAddress)
	if err != nil {
		log.Errorf("通过address从表中查询user失败， err: %v, address: %s", err, c.FromAddress)
		return "", err
	}
	if !user.CheckPassword(c.Password) {
		log.Errorf("密码有误")
		return "", utils.VerifyPasswordErr
	}

	client := transaction.NewChainClient(conf.EthChainNet, big.NewInt(int64(conf.EthChainID)))

	// 1. 检查from的usdt余额是否足够
	usdtBalance, err := client.GetTokenBalance(common.HexToAddress(c.FromAddress), common.HexToAddress(conf.EthUSDTTokenAddress))
	if err != nil {
		log.Errorf("获取eth usdt余额error: %v", err)
		return "", err
	}
	// 2. 比较usdt余额
	amount, _ := new(big.Int).SetString(c.Amount, 10)
	if usdtBalance.Cmp(amount) < 0 {
		log.Errorf("eth usdt转账余额不足；转账amount: %s, 余额：%s, address: %s", c.Amount, usdtBalance.String(), c.FromAddress)
		return "", errors.New(fmt.Sprintf("eth usdt转账余额不足；转账amount: %s, 余额：%s, address: %s", c.Amount, usdtBalance.String(), c.FromAddress))
	}

	// 3. 发送交易
	// 3.1 解码私钥
	private, err := utils.DecryptPrivate(user.PrivateCrypted)
	if err != nil {
		log.Errorf("私钥aes解码失败， error: %v, address: %s", err, user.Address)
		return "", err
	}

	// 3.2 获取suggest gasPrice
	suggestPrice, err := client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Errorf("获取suggest gasPrice error: %v", err)
		return "", err
	}
	gasLimit := uint64(60000)
	gasPrice := suggestPrice.Mul(suggestPrice, big.NewInt(2)) // 两倍suggest gasPrice

	// 3.3 获取nonce
	nonce, err := client.GetLatestNonce(c.FromAddress)
	if err != nil {
		log.Errorf("获取nonce失败, error: %v,address: %s", err, c.FromAddress)
		return "", err
	}
	// 3.4 把交易info保存到sendTransfer表
	ss := &models.SendTransfer{
		FromAddress:  c.FromAddress,
		ToAddress:    c.ToAddress,
		Amount:       c.Amount,
		TokenAddress: conf.EthUSDTTokenAddress,
		TxHash:       "",
		OwnChain:     conf.EthChainTag,
		CoinType:     3, // usdt币
		TxStatus:     0,
		ErrMsg:       "",
	}
	err = ss.Create()
	if err != nil {
		client.SetFailNonce(c.FromAddress, nonce)
		log.Errorf("创建sendTransfer记录失败； error: %v", err)
		return "", err
	}

	// 4. 发送交易
	tx, err := client.SendTokenTx(private, nonce, gasLimit, gasPrice, common.HexToAddress(c.ToAddress), common.HexToAddress(conf.EthUSDTTokenAddress), amount)
	if err != nil {
		// 交易发送失败，设置交易记录状态
		_ = ss.Update(models.SendTransfer{TxStatus: 2, ErrMsg: err.Error()})
		// 回归nonce
		client.SetFailNonce(c.FromAddress, nonce)
		log.Errorf("发送eth usdt交易失败；error: %v", err)
		return "", err
	}
	// 4.1 更新交易hash
	_ = ss.Update(models.SendTransfer{TxHash: tx.Hash().String()})
	// 4.2 保存交易到kv表中
	byteTx, _ := tx.MarshalJSON()
	_ = models.SetKv(tx.Hash().String(), byteTx)

	// 5. 监听交易链上状态
	go func() {
		count := 0
		for {
			if count > 10 {
				// 超时
				_ = ss.Update(models.SendTransfer{TxStatus: 3})
				// 回归nonce
				client.SetFailNonce(c.FromAddress, nonce)
				return
			}

			time.Sleep(15 * time.Second)
			_, isPending, err := client.Client.TransactionByHash(context.Background(), tx.Hash())
			if err == nil && !isPending {
				// 查询到了交易，修改交易状态为成功
				log.Infof("链上查询到了转账交易；txHash: %s", tx.Hash().String())
				_ = ss.Update(models.SendTransfer{TxStatus: 1})
				return
			}
			// 增加count
			count++
		}
	}()

	return tx.Hash().String(), nil
}
