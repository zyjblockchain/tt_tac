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
	"github.com/zyjblockchain/tt_tac/utils/ding_robot"
	eth_watcher "github.com/zyjblockchain/tt_tac/utils/eth-watcher"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/blockchain"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/plugin"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/structs"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"strings"
	"sync"
	"time"
)

type TacProcess struct {
	ChainNetUrl           string                       // 链的网络url
	ListenTokenAddress    string                       // 监听的代币合约地址
	TransferTokenAddress  string                       // 需要转入的合约地址
	TransferMiddleAddress string                       // 中转地址
	FromChainWatcher      *eth_watcher.AbstractWatcher // 跨链from watcher
	ToChainWatcher        *eth_watcher.AbstractWatcher // 跨链 to watcher
	lock                  sync.Mutex
}

func NewTacProcess(chainNet, listenTokenAddress, transferTokenAddress, transferMiddleAddress string, fromChainWatcher, toChainWatcher *eth_watcher.AbstractWatcher) *TacProcess {
	return &TacProcess{
		ChainNetUrl:           chainNet,
		ListenTokenAddress:    listenTokenAddress,
		TransferTokenAddress:  transferTokenAddress,
		TransferMiddleAddress: transferMiddleAddress,
		FromChainWatcher:      fromChainWatcher,
		ToChainWatcher:        toChainWatcher,
	}
}

// ListenErc20CollectionAddress 监听erc20代币收款地址
func (t *TacProcess) ListenErc20CollectionAddress() {
	callback := func(txAndReceipt *structs.RemovableTxAndReceipt) {
		events := utils.ExtractERC20TransfersIfExist(txAndReceipt)
		for _, e := range events {
			tokenAddress := e.Token
			from := e.From
			to := e.To
			amount := e.Value
			isRemoved := txAndReceipt.IsRemoved
			log.Infof("tokenAddress: %s; from: %s, to: %s, amount: %s, isRemoved: %v", tokenAddress, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String(), isRemoved)
			if (!isRemoved) && strings.ToLower(utils.FormatAddressHex(tokenAddress)) == strings.ToLower(t.ListenTokenAddress) && strings.ToLower(utils.FormatAddressHex(to)) == strings.ToLower(utils.FormatAddressHex(t.TransferMiddleAddress)) {
				log.Infof("监听到跨链转账交易：tokenAddress: %s; from: %s, to: %s, amount: %s", tokenAddress, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String())
				// 监听到转入的交易
				// 开启一个协程来执行处理此交易
				go func() {
					err := t.ProcessCollectionTx(from, amount.String())
					if err != nil {
						// 钉钉群推送
						content := fmt.Sprintf("tac 跨链转账失败；\nfrom：%s, \ntokenAddress: %s, \namount: %s. \nerror: %s", utils.FormatAddressHex(from), tokenAddress, amount.String(), err.Error())
						_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, false)
						log.Errorf("执行跨链转账逻辑失败，error: %v；from: %s; to: %s; amount: %s; tokenAddress: %s", err, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String(), tokenAddress)
					}
				}()
			}
		}
	}

	filterFunc := func(tx blockchain.Transaction) bool {
		to := tx.GetTo()
		return strings.ToLower(t.ListenTokenAddress) == strings.ToLower(to)
	}
	t.FromChainWatcher.RegisterTxReceiptPlugin(plugin.NewTxReceiptPluginWithFilter(callback, filterFunc))

	// t.FromChainWatcher.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(func(tokenAddress, from, to string, amount decimal.Decimal, isRemoved bool) {
	// 	// log.Infof("tokenAddress: %s; from: %s, to: %s, amount: %s, isRemoved: %v", tokenAddress, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String(), isRemoved)
	// 	if (!isRemoved) && strings.ToLower(utils.FormatAddressHex(tokenAddress)) == strings.ToLower(t.ListenTokenAddress) && strings.ToLower(utils.FormatAddressHex(to)) == strings.ToLower(utils.FormatAddressHex(t.TransferMiddleAddress)) {
	// 		log.Infof("监听到跨链转账交易：tokenAddress: %s; from: %s, to: %s, amount: %s", tokenAddress, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String())
	// 		// 监听到转入的交易
	// 		// 开启一个协程来执行处理此交易
	// 		go func() {
	// 			err := t.ProcessCollectionTx(from, amount.String())
	// 			if err != nil {
	// 				// 钉钉群推送
	// 				content := fmt.Sprintf("tac 跨链转账失败；\nfrom：%s, \ntokenAddress: %s, \namount: %s. \nerror: %s", utils.FormatAddressHex(from), tokenAddress, amount.String(), err.Error())
	// 				_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, false)
	// 				log.Errorf("执行跨链转账逻辑失败，error: %v；from: %s; to: %s; amount: %s; tokenAddress: %s", err, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String(), tokenAddress)
	// 			}
	// 		}()
	// 	}
	// }))
}

// ProcessCollectionTx 处理接收token逻辑
func (t *TacProcess) ProcessCollectionTx(from, amount string) error {
	if len(from) != 42 {
		from = utils.FormatAddressHex(from)
	}
	// 1. 保存监听到的转入交易到collection表中
	cc := &models.CollectionTx{
		From:         from,
		To:           t.TransferMiddleAddress,
		TokenAddress: t.ListenTokenAddress,
		Amount:       amount,
		ChainNetUrl:  t.ChainNetUrl,
		ExtraInfo:    "跨链转账",
	}
	if err := cc.Create(); err != nil {
		log.Errorf("保存collection失败：error: %v", err)
		return err
	}
	var ordType int
	if t.ChainNetUrl == conf.EthChainNet {
		ordType = conf.EthToTtOrderType
	} else if t.ChainNetUrl == conf.TTChainNet {
		ordType = conf.TtToEthOrderType
	} else {
		return errors.New("判断不了watcher监听到的orderType")
	}
	// 2. 从订单表中查询是否存在from的订单
	// ord, err := (&models.TacOrder{FromAddr: from, Amount: amount, OrderType: ordType, State: 0}).GetOrder()
	ord, exist := new(models.TacOrder).Exist(from, amount, ordType, 0)
	if !exist {
		// todo 监听到的收款信息在order中查不到，可能是充值余额到中转地址的操作，所以不用退还，需要钉钉推送通知
		content := fmt.Sprintf("tac 收到了一笔没有转账订单的转入交易；\nfrom: %s, \nto: %s, \ntokenAddress: %s, \namount: %s",
			utils.FormatAddressHex(from), utils.FormatAddressHex(t.TransferMiddleAddress), utils.FormatAddressHex(t.TransferTokenAddress), amount)
		_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, false)
		return nil
	}

	// 3. 存在则更新collection表状态，并更新order记录中的collectionId todo 后面优化成事务的方式来更新
	cc.IsValid = 1
	updateCC := models.CollectionTx{IsValid: 1}
	if err := cc.Update(updateCC); err != nil {
		log.Errorf("更新collection的valid 字段失败：%v", err)
		return err
	}
	ord.CollectionId = cc.ID
	updateOrd := models.TacOrder{CollectionId: cc.ID}
	if err := ord.Update(updateOrd); err != nil {
		log.Errorf("更新order的collectionId 字段失败：%v", err)
		return err
	}

	// 4. 执行跨链转账交易发送逻辑发送交易
	var chainNetUrl string
	var chainId int64
	var chainTag int
	if ord.OrderType == conf.EthToTtOrderType { // 以太坊上的token转到tt链上的token
		chainNetUrl = conf.TTChainNet
		chainId = conf.TTChainID
		chainTag = conf.TTChainTag
	} else if ord.OrderType == conf.TtToEthOrderType {
		chainNetUrl = conf.EthChainNet
		chainId = conf.EthChainID
		chainTag = conf.EthChainTag
	}

	// 发送token
	client := transaction.NewChainClient(chainNetUrl, big.NewInt(chainId))
	defer client.Close()

	// 4.1 发送交易之前把交易记录到tx表中，并把次记录绑定到collection上
	tt := &models.TxTransfer{
		SenderAddress:   t.TransferMiddleAddress,
		ReceiverAddress: ord.RecipientAddr,
		TokenAddress:    t.TransferTokenAddress,
		Amount:          utils.TransformAmount(amount, ord.OrderType),
		TxHash:          "",
		TxStatus:        0,
		OwnChain:        chainTag,
	}
	if err := tt.Create(); err != nil {
		log.Errorf("保存txTransfer error: %v", err)
		if err := ord.Update(models.TacOrder{State: 2}); err != nil {
			log.Errorf("修改order状态为失败状态 error: %v. orderId: %d", err, ord.ID)
		}
		return err
	}
	// 更新collection表中的TxId
	if err := cc.Update(models.CollectionTx{TxId: tt.ID}); err != nil {
		log.Errorf("更新collectionTx中的TxId失败：%v", err)
		if err := ord.Update(models.TacOrder{State: 2}); err != nil {
			log.Errorf("修改order状态为失败状态 error: %v. orderId: %d", err, ord.ID)
		}
		return err
	}
	// 4.2 发送交易
	tokenAmount, _ := new(big.Int).SetString(tt.Amount, 10)
	log.Infof("需要发送交易的tokenAmount: %v", tokenAmount)
	receiver := common.HexToAddress(tt.ReceiverAddress)
	tokenAddress := common.HexToAddress(tt.TokenAddress)

	// 查看from地址是否有足够的tokenAmount
	tokenBala, err := client.GetTokenBalance(common.HexToAddress(conf.TacMiddleAddress), tokenAddress)
	if err != nil {
		log.Errorf("获取token balance error: %v, address: %s, tokenAddress: %s", err, conf.TacMiddleAddress, tt.TokenAddress)
		if err := ord.Update(models.TacOrder{State: 2}); err != nil {
			log.Errorf("修改order状态为失败状态 error: %v. orderId: %d", err, ord.ID)
		}
		return err
	}
	if tokenBala.Cmp(tokenAmount) < 0 {
		if err := ord.Update(models.TacOrder{State: 2}); err != nil {
			log.Errorf("修改order状态为失败状态 error: %v. orderId: %d", err, ord.ID)
		}
		return errors.New("跨链转账中转地址上的token余额不足")
	}
	// gasLimit, err := client.EstimateTokenTxGas(tokenAmount, common.HexToAddress(tt.SenderAddress), tokenAddress, receiver)
	// log.Infof("预估gasLimit: %d", gasLimit)
	// if err != nil {
	// 	log.Errorf("调用预估交易gas接口失败：%v", err)
	// 	return err
	// }
	gasLimit := uint64(70000)
	suggestGasPrice, err := client.SuggestGasPrice()
	if err != nil {
		log.Errorf("获取建议的gas price失败：%v", err)
		return err
	}
	log.Infof("建议gas price: %s", suggestGasPrice.String())
	gasPrice := suggestGasPrice.Mul(suggestGasPrice, big.NewInt(2)) // 2倍的建议gas价格
	// 把 gasPrice更新到txTransfer中
	_ = tt.Update(models.TxTransfer{GasPrice: gasPrice.String()})

	nonce, err := client.GetLatestNonce(tt.SenderAddress)
	if err != nil {
		log.Errorf("获取地址nonce失败; addr: %s, error: %v", tt.SenderAddress, err)
		return err
	}
	log.Infof("跨链转账交易nonce; address: %s, nonce: %d", conf.TacMiddleAddress, nonce)
	// 组装签名交易 -> 发送上链
	signedTx, err := client.NewSignedTokenTx(conf.TacMiddleAddressPrivate, nonce, gasLimit, gasPrice, receiver, tokenAddress, tokenAmount)
	if err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(tt.SenderAddress, nonce)
		// 修改TxTransfer表中的状态置位失败，并存入失败信息
		if err := tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()}); err != nil {
			log.Errorf("更新TxTransfer交易失败状态到数据库失败：error: %v", err)
			return err
		}
		if err := ord.Update(models.TacOrder{State: 2}); err != nil {
			log.Errorf("修改order状态为失败状态 error: %v. orderId: %d", err, ord.ID)
		}
		return err
	}
	// 4.3 把签名交易更新TxTransfer中的txHash,并存储在kv表中 todo 事务处理
	if err := tt.Update(models.TxTransfer{TxHash: signedTx.Hash().String()}); err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(tt.SenderAddress, nonce)

		log.Errorf("更新TxTransfer交易hash到数据库失败：error: %v", err)
		return err
	}
	byteTx, err := signedTx.MarshalJSON()
	if err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(tt.SenderAddress, nonce)
		log.Errorf("marshal tx err: %v", err)
		return err
	}
	if err := models.SetKv(signedTx.Hash().String(), byteTx); err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(tt.SenderAddress, nonce)
		log.Errorf("把交易保存到kv表中失败：%v", err)
		return err
	}
	// 4.4 发送交易上链
	if err := client.Client.SendTransaction(context.Background(), signedTx); err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(tt.SenderAddress, nonce)
		log.Errorf("发送签好名的交易上链失败；error: %v", err)
		if err := tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()}); err != nil {
			log.Errorf("更新TxTransfer交易失败状态到数据库失败：error: %v", err)
			return err
		}
		if err := ord.Update(models.TacOrder{State: 2}); err != nil {
			log.Errorf("修改order状态为失败状态 error: %v. orderId: %d", err, ord.ID)
		}
		return err
	}
	// 更新发送交易的hash到tacOrder的ReceiveTxHash上
	_ = ord.Update(models.TacOrder{ReceiveTxHash: signedTx.Hash().String()}) // todo 不是很重要的更新字段  忽略error
	log.Infof("tac 发送交易上链；txHash: %s", signedTx.Hash().String())

	// 5. 注册监听刚发送的交易状态
	t.lock.Lock() // 保证注册交易监听的线程安全
	defer t.lock.Unlock()
	timeoutTimestamp := time.Now().Add(10 * time.Second).Unix() // 首次超时时间为10s
	count := 1
	pluginIndex := len(t.ToChainWatcher.TxPlugins) // todo 线程不安全
	t.ToChainWatcher.RegisterTxPlugin(plugin.NewTxHashPlugin(func(txHash string, isRemoved bool) {
		if strings.ToLower(signedTx.Hash().String()) == strings.ToLower(txHash) {
			// 监听到此交易
			log.Infof("链上监听到成功发送的跨链转账交易；txHash: %s", txHash)
			// 1. 修改交易状态为成功 todo 事务更新
			if err := tt.Update(models.TxTransfer{TxStatus: 1}); err != nil {
				log.Errorf("修改交易状态为success error: %v. txHash: %s", err, txHash)
			}
			// 2. 修改订单状态为完成
			if err := ord.Update(models.TacOrder{State: 1}); err != nil {
				log.Errorf("修改order状态为success error: %v. orderId: %d", err, ord.ID)
			}
			// 注销此监听
			t.ToChainWatcher.UnRegisterTxPlugin(pluginIndex)
			return
		}
		// 判断监听是否超时，超时则重新发送交易到链上(因为tt链存在分叉把交易丢失的情况)
		now := time.Now().Unix()
		if now > timeoutTimestamp {
			// 重发次数大于5次则失败
			if count > 5 {
				// 把获取的address nonce 置为 fail
				client.SetFailNonce(tt.SenderAddress, nonce)
				log.Errorf("跨链转账交易监听超时； txHash: %s", txHash)
				// 修改交易状态为超时 todo 事务更新
				if err := tt.Update(models.TxTransfer{TxStatus: 3}); err != nil {
					log.Errorf("修改交易状态为超时error: %v. txHash: %s", err, txHash)
				}
				if err := ord.Update(models.TacOrder{State: 3}); err != nil {
					log.Errorf("修改order状态为超时 error: %v. orderId: %d", err, ord.ID)
				}
				// 注销此监听
				t.ToChainWatcher.UnRegisterTxPlugin(pluginIndex)
				return
			}

			// 重新发送一次交易到链上
			_ = client.Client.SendTransaction(context.Background(), signedTx)
			log.Infof("tac 交易监听时间超过了超时时间，重新发送交易到链上；txHash: %s", signedTx.Hash().String())
			count++
			// 重置超时时间,累加10s
			t := time.Duration(count * 20)
			timeoutTimestamp = time.Now().Add(t * time.Second).Unix()
		}
	}))
	return nil
}
