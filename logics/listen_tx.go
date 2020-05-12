package logics

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
	eth_watcher "github.com/zyjblockchain/tt_tac/utils/eth-watcher"
	"github.com/zyjblockchain/tt_tac/utils/eth-watcher/plugin"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
)

type TacProcess struct {
	ChainNet        string // 链的网络
	TokenAddress    string // 代币合约地址
	TransferAddress string // 中转地址
}

// 监听erc20代币收款地址
func (t TacProcess) ListenErc20CollectionAddress() {
	w := eth_watcher.NewHttpBasedEthWatcher(context.Background(), t.ChainNet)
	w.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(func(tokenAddress, from, to string, amount decimal.Decimal, isRemoved bool) {
		if tokenAddress == t.TokenAddress && utils.FormatHex(to) == utils.FormatHex(t.TransferAddress) {
			// 监听到转入的交易
			// 开启一个协程来执行处理此交易 todo

		}
	}))
	if err := w.RunTillExit(); err != nil {
		log.Errorf("监听链上交易失败：%v", err)
	}
}

// ProcessCollectionTx 处理接收token逻辑
func (t TacProcess) ProcessCollectionTx(from, amount string) error {
	if len(from) != 42 {
		from = utils.FormatHex(from)
	}
	// 从订单表中查询是否存在from的订单
	ord, err := models.Order{FromAddr: from}.GetByAddr()
	if err != nil {
		log.Errorf("通过fromAddr查询订单表失败：from: %s, err: %v", from, err)
		return err
	}
	// 存在该from订单则发送交易
	tokenAmount, ok := new(big.Int).SetString(ord.Amount, 10)
	if !ok {
		log.Errorf("string转big Int失败")
		return errors.New("string转big Int失败")
	}
	var chainNetUrl string
	var chainId int64
	if ord.OrderType == models.EthToTtOrderType { // 以太坊上的token转到tt链上的token
		chainNetUrl = conf.TTMainNet
		chainId = conf.TTMainNetID
	} else if ord.OrderType == models.TtToEthOrderType {
		chainNetUrl = conf.EthNet
		chainId = conf.EthNetID
	}
	// 发送token
	client := transaction.NewChainClient(chainNetUrl, big.NewInt(chainId))
	defer client.Close()

	// 发送交易之前把交易记录到tx表中 todo 明天continue

}
