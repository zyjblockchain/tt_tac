package logics

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
	"github.com/zyjblockchain/tt_tac/utils/ding_robot"
	eth_watcher "github.com/zyjblockchain/tt_tac/utils/eth-watcher"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/plugin"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"strings"
	"sync"
)

// 闪兑 usdt -> eth_pala
type FlashChange struct {
	FromTokenAddress string `json:"from_token_address"`                   // usdt token address，目前是写死，后面为了扩展会要求前端传
	ToTokenAddress   string `json:"to_token_address"`                     // eth_pala token address，目前是写死，后面为了扩展会要求前端传
	OperateAddress   string `json:"operate_address" binding:"required"`   // 操作地址
	FromTokenAmount  string `json:"from_token_amount" binding:"required"` // 需要兑换数量,usdt的数量
	ToTokenAmount    string `json:"to_token_amount"`                      // 转换之后的数量, usdt按照实时价格转换pala的数量。todo 为收取手续费，所以实时价格需要提高1%进行兑换
}

func (f *FlashChange) FlashChange() (string, error) {
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
		return "", errors.New(fmt.Sprintf("存在正在进行中的同类型的闪兑订单。operateAddress: %s, fromToken: %s, toToken: %s", f.OperateAddress, f.FromTokenAddress, f.ToTokenAddress))
	}

	// 2. 查看operateAddress是否有足够多的FromTokenAmount
	client := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer client.Close()
	fromTokenBalance, err := client.GetTokenBalance(common.HexToAddress(f.OperateAddress), common.HexToAddress(f.FromTokenAddress))
	if err != nil {
		log.Errorf("get token balance error: %v, address: %s, tokenAddress: %s", err, f.OperateAddress, f.FromTokenAddress)
		return "", err
	}
	fromTokenAmount, _ := new(big.Int).SetString(f.FromTokenAmount, 10)
	if fromTokenBalance.Cmp(fromTokenAmount) < 0 {
		// 余额不足
		log.Errorf("账户中闪兑的from token 余额不足。address: %s, tokenAddress: %s, fromTokenBalance：%s, fromTokenAmount: %s",
			f.OperateAddress, f.FromTokenAddress, fromTokenBalance.String(), f.FromTokenAmount)
		return "", errors.New(fmt.Sprintf("账户中闪兑的from token 余额不足。address: %s, tokenAddress: %s, fromTokenBalance：%s, fromTokenAmount: %s",
			f.OperateAddress, f.FromTokenAddress, fromTokenBalance.String(), f.FromTokenAmount))
	}

	// todo 获取兑换价格，赋值给ToTokenAmount，
	//  这里前端会传的，但是为了防止前端传入错误，这里需要再次获取来计算一次兑换数量

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
		return "", err
	}
	// 3.2 发送交易闪兑的usdt到闪兑的中转账户
	from := f.OperateAddress
	u, err := new(models.User).GetUserByAddress(f.OperateAddress)
	if err != nil {
		log.Errorf("通过address从表中查询user失败， err: %v, address: %s", err, f.OperateAddress)
		return "", err
	}
	fromPrivate, err := utils.DecryptPrivate(u.PrivateCrypted)
	if err != nil {
		log.Errorf("aes 解码私钥失败。 err：%v", err)
		return "", err
	}
	to := common.HexToAddress(conf.EthFlashChangeMiddleAddress)
	tokenAddress := common.HexToAddress(f.FromTokenAddress)
	tokenAmount, _ := new(big.Int).SetString(f.FromTokenAmount, 10)
	suggestPrice, err := client.SuggestGasPrice()
	if err != nil {
		log.Errorf("get suggest gasPrice err : %v", err)
		return "", err
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
		return "", err
	}
	// 3.4 生成签名交易
	signedTx, err := client.NewSignedTokenTx(fromPrivate, nonce, gasLimit, gasPrice, to, tokenAddress, tokenAmount)
	if err != nil {
		log.Errorf("发送交易前对交易组装签名失败。error: %v", err)
		// 更新交易状态为失败
		_ = tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()})
		return "", err
	}
	// 3.5 更新交易hash到txTransfer中
	if err := tt.Update(models.TxTransfer{TxHash: signedTx.Hash().String()}); err != nil {
		log.Errorf("更新交易hash 到TxTransfer失败。error: %v, tx: %v", err, *signedTx)
		return "", err
	}
	// 3.6 把交易保存到kv表中
	byteTx, err := signedTx.MarshalJSON()
	if err != nil {
		log.Errorf("marshal tx err: %v", err)
		return "", err
	}
	if err := models.SetKv(signedTx.Hash().String(), byteTx); err != nil {
		log.Errorf("把交易保存到kv表中失败：%v", err)
		return "", err
	}
	// 3.7 发送交易上链
	if err := client.Client.SendTransaction(context.Background(), signedTx); err != nil {
		log.Errorf("发送签好名的交易上链失败。 error: %v, txHash: %s", err, signedTx.Hash().String())
		_ = tt.Update(models.TxTransfer{TxStatus: 2, ErrMsg: err.Error()})
		return "", err
	}
	// 3.8 更新FlashChangeOrder 上的sendTxId
	if err := fo.Update(models.FlashChangeOrder{SendTxId: tt.ID}); err != nil {
		log.Errorf("更新 FlashChangeOrder 上的sendTxId error: %v", err)
		return "", err
	}
	return signedTx.Hash().String(), nil
}

// 开启一个wather来监听闪兑的交易
type WatchFlashChange struct {
	FromTokenAddress string
	ToTokenAddress   string
	ChainWatcher     *eth_watcher.AbstractWatcher
	lock             sync.Mutex
}

func (w *WatchFlashChange) ListenFlashChangeTx() {
	w.ChainWatcher.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(func(tokenAddress, from, to string, amount decimal.Decimal, isRemoved bool) {
		// 监听中转地址接收到usdt的交易
		if (!isRemoved) && strings.ToLower(utils.FormatAddressHex(tokenAddress)) == strings.ToLower(w.FromTokenAddress) && strings.ToLower(utils.FormatAddressHex(to)) == strings.ToLower(utils.FormatAddressHex(conf.EthFlashChangeMiddleAddress)) {
			log.Infof("监听到闪兑接收地址有转入记录：tokenAddress: %s; from: %s, to: %s, amount: %s", tokenAddress, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String())
			// 开启一个协程来处理闪兑接收地址
			go func() {
				err := w.processCollectFlashChangeTx(utils.FormatAddressHex(from), amount.String())
				if err != nil {
					// 钉钉群推送
					content := fmt.Sprintf("闪兑失败；\nfrom：%s, \ntokenAddress: %s, \namount: %s", utils.FormatAddressHex(from), tokenAddress, amount.String())
					_ = ding_robot.NewRobot(utils.WebHook).SendText(content, nil, true)
					log.Errorf("执行闪兑逻辑失败，error: %v；from: %s; to: %s; amount: %s; tokenAddress: %s", err, utils.FormatAddressHex(from), utils.FormatAddressHex(to), amount.String(), tokenAddress)
				}
			}()
		}
	}))
}

// processCollectFlashChangeTx 处理监听的闪兑接口监听到转入交易流程
func (w *WatchFlashChange) processCollectFlashChangeTx(from, amount string) error {

}
