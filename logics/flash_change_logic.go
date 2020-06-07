package logics

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
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

// 闪兑 usdt -> eth_pala
type FlashChange struct {
	Password         string `json:"password" binding:"required"`          // 支付密码
	FromTokenAddress string `json:"from_token_address"`                   // usdt token address，目前是写死，后面为了扩展会要求前端传
	ToTokenAddress   string `json:"to_token_address"`                     // eth_pala token address，目前是写死，后面为了扩展会要求前端传
	OperateAddress   string `json:"operate_address" binding:"required"`   // 操作地址
	FromTokenAmount  string `json:"from_token_amount" binding:"required"` // 需要兑换数量,usdt的数量
	ToTokenAmount    string `json:"to_token_amount"`                      // 转换之后的数量, usdt按照实时价格转换pala的数量。todo 为收取手续费，所以实时价格需要提高1%进行兑换
	TradePrice       string `json:"trade_price" binding:"required"`       // 兑换价格，目前是pala_usdt价格
}

// FlashChange 返回闪兑的订单号
func (f *FlashChange) FlashChange() (uint, error) {
	// 0. 验证支付密码
	user, err := new(models.User).GetUserByAddress(f.OperateAddress)
	if err != nil {
		log.Errorf("通过address从表中查询user失败， err: %v, address: %s", err, f.OperateAddress)
		return 0, err
	}
	if !user.CheckPassword(f.Password) {
		log.Errorf("密码有误")
		return 0, utils.VerifyPasswordErr
	}
	// 1. 查看operateAddress 是否存在正在进行中的闪兑订单
	exist := new(models.FlashChangeOrder).Exist(f.OperateAddress, f.FromTokenAddress, f.ToTokenAddress, 0)
	if exist {
		// 存在则返回
		log.Errorf("存在正在进行中的同类型的闪兑订单。operateAddress: %s, fromToken: %s, toToken: %s", f.OperateAddress, f.FromTokenAddress, f.ToTokenAddress)
		return 0, errors.New(fmt.Sprintf("存在正在进行中的同类型的闪兑订单。operateAddress: %s, fromToken: %s, toToken: %s", f.OperateAddress, f.FromTokenAddress, f.ToTokenAddress))
	}

	// 2. 查看operateAddress是否有足够多的FromTokenAmount
	client := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer client.Close()
	fromTokenBalance, err := client.GetTokenBalance(common.HexToAddress(f.OperateAddress), common.HexToAddress(f.FromTokenAddress))
	if err != nil {
		log.Errorf("get token balance error: %v, address: %s, tokenAddress: %s", err, f.OperateAddress, f.FromTokenAddress)
		return 0, err
	}
	fromTokenAmount, _ := new(big.Int).SetString(f.FromTokenAmount, 10)
	if fromTokenBalance.Cmp(fromTokenAmount) < 0 {
		// 余额不足
		log.Errorf("账户中闪兑的from token 余额不足。address: %s, tokenAddress: %s, fromTokenBalance：%s, fromTokenAmount: %s",
			f.OperateAddress, f.FromTokenAddress, fromTokenBalance.String(), f.FromTokenAmount)
		return 0, errors.New(fmt.Sprintf("账户中闪兑的from token 余额不足。address: %s, tokenAddress: %s, fromTokenBalance：%s, fromTokenAmount: %s",
			f.OperateAddress, f.FromTokenAddress, fromTokenBalance.String(), f.FromTokenAmount))
	}

	// todo 获取兑换价格，赋值给ToTokenAmount，
	//  这里前端会传的，但是为了防止前端传入错误，这里需要再次获取来计算一次兑换数量

	// 3. 执行闪兑process
	// 3.1 保存闪兑订单
	fo := &models.FlashChangeOrder{
		OperateAddress:   strings.ToLower(f.OperateAddress),
		FromTokenAddress: f.FromTokenAddress,
		ToTokenAddress:   f.ToTokenAddress,
		FromTokenAmount:  f.FromTokenAmount,
		ToTokenAmount:    f.ToTokenAmount,
		TradePrice:       f.TradePrice,
		State:            0,
	}
	if err := fo.Create(); err != nil {
		log.Errorf("保存flash change order error: %v", err)
		return 0, err
	}
	// 3.2 发送交易闪兑的usdt到闪兑的中转账户
	from := f.OperateAddress
	// 获取私钥
	fromPrivate, err := utils.DecryptPrivate(user.PrivateCrypted)
	if err != nil {
		log.Errorf("aes 解码私钥失败。 err：%v", err)
		return 0, err
	}
	to := common.HexToAddress(conf.EthFlashChangeMiddleAddress)
	tokenAddress := common.HexToAddress(f.FromTokenAddress)
	tokenAmount, _ := new(big.Int).SetString(f.FromTokenAmount, 10)
	suggestPrice, err := client.SuggestGasPrice()
	if err != nil {
		log.Errorf("get suggest gasPrice err : %v", err)
		return 0, err
	}
	gasPrice := suggestPrice.Mul(suggestPrice, big.NewInt(2)) // 两倍于suggest gasPrice
	gasLimit := uint64(60000)
	nonce, err := client.GetNonce(common.HexToAddress(from))
	if err != nil {
		log.Errorf("get nonce error: %v, address: %s", err, from)
	}
	// 3.3 把交易存入tx_transfer表中
	tt := &models.TxTransfer{
		SenderAddress:   from,
		ReceiverAddress: conf.EthFlashChangeMiddleAddress,
		TokenAddress:    f.FromTokenAddress,
		Amount:          f.FromTokenAmount,
		GasPrice:        gasPrice.String(),
		TxStatus:        0,
		OwnChain:        conf.EthChainTag,
	}
	if err := tt.Create(); err != nil {
		log.Errorf("保存交易到TxTransfer失败。 error: %v", err)
		return 0, err
	}
	// 3.4 生成签名交易
	signedTx, err := client.NewSignedTokenTx(fromPrivate, nonce, gasLimit, gasPrice, to, tokenAddress, tokenAmount)
	if err != nil {
		log.Errorf("发送交易前对交易组装签名失败。error: %v", err)
		// 更新交易状态为失败
		_ = tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()})
		return 0, err
	}
	// 3.5 更新交易hash到txTransfer中
	if err := tt.Update(models.TxTransfer{TxHash: signedTx.Hash().String()}); err != nil {
		log.Errorf("更新交易hash 到TxTransfer失败。error: %v, tx: %v", err, *signedTx)
		return 0, err
	}
	// 3.6 把交易保存到kv表中
	byteTx, err := signedTx.MarshalJSON()
	if err != nil {
		log.Errorf("marshal tx err: %v", err)
		return 0, err
	}
	if err := models.SetKv(signedTx.Hash().String(), byteTx); err != nil {
		log.Errorf("把交易保存到kv表中失败：%v", err)
		return 0, err
	}
	// 3.7 发送交易上链
	if err := client.Client.SendTransaction(context.Background(), signedTx); err != nil {
		log.Errorf("发送签好名的交易上链失败。 error: %v, txHash: %s", err, signedTx.Hash().String())
		_ = tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()})
		return 0, err
	}
	// from记录进map中
	FlashAddressMap[strings.ToLower(from)] = 1

	// 3.8 更新FlashChangeOrder 上的sendTxId
	if err := fo.Update(models.FlashChangeOrder{SendTxId: tt.ID}); err != nil {
		log.Errorf("更新 FlashChangeOrder 上的sendTxId error: %v", err)
		return 0, err
	}
	return fo.ID, nil
}

var FlashAddressMap = make(map[string]int)

// 开启一个wather来监听闪兑的交易
type WatchFlashChange struct {
	FromTokenAddress string // usdt token address
	ToTokenAddress   string // pala token address
	ChainWatcher     *eth_watcher.AbstractWatcher
	lock             sync.Mutex
}

func NewWatchFlashChange(fromTokenAddress, toTokenAddress string, ethChainWatcher *eth_watcher.AbstractWatcher) *WatchFlashChange {
	return &WatchFlashChange{
		FromTokenAddress: fromTokenAddress,
		ToTokenAddress:   toTokenAddress,
		ChainWatcher:     ethChainWatcher,
		lock:             sync.Mutex{},
	}
}

func (w *WatchFlashChange) ListenFlashChangeTx() {
	callback := func(txAndReceipt *structs.RemovableTxAndReceipt) {
		events := utils.ExtractERC20TransfersIfExist(txAndReceipt)
		for _, e := range events {
			tokenAddress := e.Token
			from := e.From
			to := e.To
			amount := e.Value
			isRemoved := txAndReceipt.IsRemoved
			// 监听中转地址接收到usdt的交易
			if (!isRemoved) && strings.ToLower(utils.FormatAddressHex(tokenAddress)) == strings.ToLower(w.FromTokenAddress) && strings.ToLower(utils.FormatAddressHex(to)) == strings.ToLower(utils.FormatAddressHex(conf.EthFlashChangeMiddleAddress)) {
				log.Infof("监听到闪兑接收地址有转入记录：tokenAddress: %s; from: %s, to: %s, amount: %s", tokenAddress, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String())
				// 开启一个协程来处理闪兑接收地址
				go func() {
					err := w.ProcessCollectFlashChangeTx(utils.FormatAddressHex(from), amount.String())
					if err != nil {
						// 钉钉群推送
						content := fmt.Sprintf("闪兑失败；\nfrom：%s, \ntokenAddress: %s, \namount: %s,\nerror: %s", utils.FormatAddressHex(from), tokenAddress, amount.String(), err.Error())
						_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, true)
						log.Errorf("执行闪兑逻辑失败，error: %v；from: %s; to: %s; amount: %s; tokenAddress: %s", err, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String(), tokenAddress)
					}
				}()
			}

		}
	}

	filterFunc := func(tx blockchain.Transaction) bool {
		to := tx.GetTo()
		if strings.ToLower(w.FromTokenAddress) == strings.ToLower(to) {
			// 判断from是否存在于闪兑的订单中
			from := tx.GetFrom()
			_, ok := FlashAddressMap[strings.ToLower(from)]
			if ok {
				log.Infof("监听到闪兑交易: %v", ok)
				// delete(FlashAddressMap, strings.ToLower(from))
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	}

	w.ChainWatcher.RegisterTxReceiptPlugin(plugin.NewTxReceiptPluginWithFilter(callback, filterFunc))
	// w.ChainWatcher.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(func(tokenAddress, from, to string, amount decimal.Decimal, isRemoved bool) {
	// 	// 监听中转地址接收到usdt的交易
	// 	if (!isRemoved) && strings.ToLower(utils.FormatAddressHex(tokenAddress)) == strings.ToLower(w.FromTokenAddress) && strings.ToLower(utils.FormatAddressHex(to)) == strings.ToLower(utils.FormatAddressHex(conf.EthFlashChangeMiddleAddress)) {
	// 		log.Infof("监听到闪兑接收地址有转入记录：tokenAddress: %s; from: %s, to: %s, amount: %s", tokenAddress, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String())
	// 		// 开启一个协程来处理闪兑接收地址
	// 		go func() {
	// 			err := w.ProcessCollectFlashChangeTx(utils.FormatAddressHex(from), amount.String())
	// 			if err != nil {
	// 				// 钉钉群推送
	// 				content := fmt.Sprintf("闪兑失败；\nfrom：%s, \ntokenAddress: %s, \namount: %s,\nerror: %s", utils.FormatAddressHex(from), tokenAddress, amount.String(), err.Error())
	// 				_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, true)
	// 				log.Errorf("执行闪兑逻辑失败，error: %v；from: %s; to: %s; amount: %s; tokenAddress: %s", err, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String(), tokenAddress)
	// 			}
	// 		}()
	// 	}
	// }))
}

// ProcessCollectFlashChangeTx 处理监听的闪兑接口监听到转入交易流程
func (w *WatchFlashChange) ProcessCollectFlashChangeTx(from, amount string) error {
	if len(from) != 42 {
		from = utils.FormatAddressHex(from)
	}
	// 1. 保存监听到的转入交易到collection表中
	cc := &models.CollectionTx{
		From:         from,
		To:           utils.FormatAddressHex(conf.EthFlashChangeMiddleAddress), // 闪兑的中转地址
		TokenAddress: w.FromTokenAddress,
		Amount:       amount,
		ChainNetUrl:  conf.EthChainNet,
		IsValid:      0,
		ExtraInfo:    "闪兑",
	}
	if err := cc.Create(); err != nil {
		log.Errorf("保存collection失败：error: %v", err)
		return err
	}

	// 2. 在FlashChangeOrder 表中查看此from是否申请了闪兑的订单
	fo, err := (&models.FlashChangeOrder{OperateAddress: strings.ToLower(from)}).Get()
	if err != nil {
		log.Errorf("通过OperateAddress查询闪兑订单失败。OperateAddress: %s, error: %v", from, err)
		content := fmt.Sprintf("tac 收到了一笔没有转账订单的USDT转入交易；\nfrom: %s, \nto: %s, \ntokenAddress: %s, \namount: %s",
			utils.FormatAddressHex(from), utils.FormatAddressHex(conf.EthFlashChangeMiddleAddress), utils.FormatAddressHex(w.FromTokenAddress), amount)
		_ = ding_robot.NewRobot(conf.AbnormalWebHook).SendText(content, nil, true)
		return nil
	}

	// 3. 存在闪兑的订单则更新
	if err := cc.Update(models.CollectionTx{IsValid: 1}); err != nil {
		log.Errorf("更新collection的valid 字段失败：%v", err)
		return err
	}
	if err := fo.Update(models.FlashChangeOrder{CollectionId: cc.ID}); err != nil {
		log.Errorf("更新FlashChangeOrder的collectionId 字段失败：%v", err)
		return err
	}
	// 更新闪兑的第一笔发送usdt交易的状态为成功
	_ = (&models.TxTransfer{Model: gorm.Model{ID: fo.SendTxId}}).Update(models.TxTransfer{TxStatus: 1})

	// 4. 执行闪兑逻辑的发送闪兑的pala交易
	client := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer client.Close()

	// 4.1 组装交易
	sender := conf.EthFlashChangeMiddleAddress
	senderPrivate := conf.EthFlashChangeMiddlePrivate
	palaTokenAddress := fo.ToTokenAddress
	receiver := fo.OperateAddress
	palaAmount, _ := new(big.Int).SetString(fo.ToTokenAmount, 10)
	log.Infof("闪兑需要发送的pala amount: %s, receiver: %s", palaAmount, receiver)

	// 判断token余额是否足够
	tokenBala, err := client.GetTokenBalance(common.HexToAddress(sender), common.HexToAddress(palaTokenAddress))
	if err != nil {
		log.Errorf("获取token balance error: %v, address: %s, tokenAddress: %s", err, conf.TacMiddleAddress, palaTokenAddress)
		return err
	}
	if tokenBala.Cmp(palaAmount) < 0 {
		return errors.New("闪兑中转地址上的token余额不足")
	}

	gasLimit := uint64(70000)
	suggestGasPrice, err := client.SuggestGasPrice()
	if err != nil {
		log.Errorf("获取建议的gas price失败：%v", err)
		return err
	}
	gasPrice := suggestGasPrice.Mul(suggestGasPrice, big.NewInt(2))

	// 4.2 发送交易之前把交易记录到tx表中，并把次记录绑定到collection上
	tt := &models.TxTransfer{
		SenderAddress:   sender,
		ReceiverAddress: receiver,
		TokenAddress:    palaTokenAddress,
		Amount:          palaAmount.String(),
		TxHash:          "",
		GasPrice:        gasPrice.String(),
		TxStatus:        0,
		OwnChain:        conf.EthChainTag,
	}
	if err := tt.Create(); err != nil {
		log.Errorf("保存txTransfer error: %v", err)
		return err
	}
	// 更新collection表中的TxId
	if err := cc.Update(models.CollectionTx{TxId: tt.ID}); err != nil {
		log.Errorf("更新collectionTx中的TxId失败：%v", err)
		return err
	}
	// 更新FlashChangeOrder 中的ReceiveTxId
	if err := fo.Update(models.FlashChangeOrder{ReceiveTxId: tt.ID}); err != nil {
		log.Errorf("更新FlashChangeOrder中的ReceiveTxId失败：%v", err)
		return err
	}

	// 获取nonce
	nonce, err := client.GetLatestNonce(sender)
	if err != nil {
		log.Errorf("获取地址nonce失败; addr: %s, error: %v", sender, err)
		return err
	}
	log.Infof("闪兑send交易nonce; address: %s, nonce: %d", sender, nonce)
	// 4.3 生成签名交易
	signedTx, err := client.NewSignedTokenTx(senderPrivate, nonce, gasLimit, gasPrice, common.HexToAddress(receiver), common.HexToAddress(palaTokenAddress), palaAmount)
	if err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(sender, nonce)
		// 修改TxTransfer表中的状态置位失败，并存入失败信息
		if err := tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()}); err != nil {
			log.Errorf("更新TxTransfer交易失败状态到数据库失败：error: %v", err)
			return err
		}
		if err := fo.Update(models.FlashChangeOrder{State: 2}); err != nil {
			log.Errorf("修改FlashChangeOrder状态为失败状态 error: %v. FlashChangeOrderId: %d", err, fo.ID)
		}
		return err
	}
	// 4.4 把签名交易更新TxTransfer中的txHash,并存储在kv表中 todo 事务处理
	if err := tt.Update(models.TxTransfer{TxHash: signedTx.Hash().String()}); err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(sender, nonce)
		log.Errorf("更新TxTransfer交易hash到数据库失败：error: %v", err)
		return err
	}
	byteTx, err := signedTx.MarshalJSON()
	if err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(sender, nonce)
		log.Errorf("marshal tx err: %v", err)
		return err
	}
	if err := models.SetKv(signedTx.Hash().String(), byteTx); err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(sender, nonce)
		log.Errorf("把交易保存到kv表中失败：%v", err)
		return err
	}
	// 4.5 发送交易上链
	if err := client.Client.SendTransaction(context.Background(), signedTx); err != nil {
		// 把获取的address nonce 置为 fail
		client.SetFailNonce(sender, nonce)
		log.Errorf("发送签好名的交易上链失败；error: %v", err)
		if err := tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()}); err != nil {
			log.Errorf("更新TxTransfer交易失败状态到数据库失败：error: %v", err)
			return err
		}
		if err := fo.Update(models.FlashChangeOrder{State: 2}); err != nil {
			log.Errorf("修改FlashChangeOrder状态为失败状态 error: %v. orderId: %d", err, fo.ID)
		}
		return err
	}
	log.Infof("闪兑 发送交易上链；txHash: %s", signedTx.Hash().String())

	// 5. 注册监听刚发送的交易状态
	w.lock.Lock() // 保证注册交易监听的线程安全
	defer w.lock.Unlock()
	timeoutTimestamp := time.Now().Add(30 * time.Second).Unix() // 首次超时时间为30s
	count := 1
	pluginIndex := len(w.ChainWatcher.TxPlugins)
	w.ChainWatcher.RegisterTxPlugin(plugin.NewTxHashPlugin(func(txHash string, isRemoved bool) {
		if strings.ToLower(signedTx.Hash().String()) == strings.ToLower(txHash) {
			// 监听到此交易
			log.Infof("链上监听到成功发送的闪兑发送交易；txHash: %s", txHash)
			// 1. 修改交易状态为成功 todo 事务更新
			if err := tt.Update(models.TxTransfer{TxStatus: 1}); err != nil {
				log.Errorf("修改交易状态为success error: %v. txHash: %s", err, txHash)
			}
			// 2. 修改订单状态为完成
			if err := fo.Update(models.FlashChangeOrder{State: 1}); err != nil {
				log.Errorf("修改FlashChangeOrder状态为success error: %v. orderId: %d", err, fo.ID)
			}
			// 注销此监听
			w.ChainWatcher.UnRegisterTxPlugin(pluginIndex)
			return
		}
		// 判断监听是否超时，超时则注销
		now := time.Now().Unix()
		if now > timeoutTimestamp {
			// 重发次数大于5次则失败
			if count > 5 {
				// 把获取的address nonce 置为 fail
				client.SetFailNonce(sender, nonce)
				log.Errorf("跨链转账交易监听超时； txHash: %s", txHash)
				// 修改交易状态为超时 todo 事务更新
				if err := tt.Update(models.TxTransfer{TxStatus: 3}); err != nil {
					log.Errorf("修改交易状态为超时error: %v. txHash: %s", err, txHash)
				}
				if err := fo.Update(models.FlashChangeOrder{State: 3}); err != nil {
					log.Errorf("修改FlashChangeOrder状态为超时 error: %v. orderId: %d", err, fo.ID)
				}
				// 注销此监听
				w.ChainWatcher.UnRegisterTxPlugin(pluginIndex)
				return
			}

			// 重新发送一次交易到链上
			_ = client.Client.SendTransaction(context.Background(), signedTx)
			log.Infof("闪兑 交易监听时间超过了超时时间，重新发送交易到链上；txHash: %s", signedTx.Hash().String())
			count++

			// 重置超时时间,累加30s
			t := time.Duration(count * 30)
			timeoutTimestamp = time.Now().Add(t * time.Second).Unix()
		}
	}))
	return nil
}
