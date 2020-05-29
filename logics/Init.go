package logics

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"time"
)

// InitTacOrderState 服务重启之后初始化跨链转账的订单状态和闪兑的订单状态
func InitTacOrderState(ethToTtProcess, ttToEthProcess *TacProcess) {
	ttClient := transaction.NewChainClient(conf.TTChainNet, big.NewInt(conf.TTChainID))
	defer ttClient.Close()
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()
	// 获取tacOrder表中状态为0的记录
	tacOrds, err := new(models.TacOrder).GetTacOrdersByState(0)
	if err != nil {
		log.Errorf("获取tacOrder表中状态为0的记录失败；error: %v", err)
		panic(err)
	}
	log.Infof("开始遍历查询出来的tac order。订单数量：%d", len(tacOrds))
	for _, ord := range tacOrds {
		// 设置process
		process := ethToTtProcess // 默认值
		if ord.OrderType == conf.EthToTtOrderType {

		} else if ord.OrderType == conf.TtToEthOrderType {
			process = ttToEthProcess
		} else {
			// 错误的订单，直接删除
			_ = ord.Delete(ord.ID)
			continue
		}

		// 1. 如果没有send_Tx_hash，则说明跨链转账申请者转账到中间地址的交易都失败了
		if ord.SendTxHash == "" {
			// 直接删除订单
			_ = ord.Delete(ord.ID)
			continue
		}
		// 2. 存在SendTxHash，则在链上查询交易是否发送成功
		// 2.1 设置链的client
		var sendClient, receiveClient *transaction.ChainClient
		// 链上查询交易hash是否存在
		if ord.OrderType == conf.EthToTtOrderType {
			sendClient = ethClient
			receiveClient = ttClient
		} else if ord.OrderType == conf.TtToEthOrderType {
			sendClient = ttClient
			receiveClient = ethClient
		} else {
			// orderType不存在则直接删除订单
			log.Errorf("该跨链转账订单的orderType不正确。orderId: %d, orderType: %d", ord.ID, ord.OrderType)
			_ = ord.Delete(ord.ID)
			continue
		}
		// 2.2 通过交易hash查询交易是否上链发送成功
		_, isPending, err := sendClient.Client.TransactionByHash(context.Background(), common.HexToHash(ord.SendTxHash))
		if err != nil { // todo 此处应该判断err ==  ethereum.NotFound,但是我们默认为我们发送的交易只要执行都会被执行成功的，也有可能网络错误，但是这里我们不考虑这种情况了
			// 交易没有发送成功，则delete记录
			log.Errorf("跨链转账申请者send交易没有上链")
			_ = ord.Delete(ord.ID)
			continue
		} else if isPending { // 这种情况可能性很少
			// 2.3 此交易正在pending
			// 开启协程轮询监听交易情况
			go func() {
				var count = 5
				for {
					count--
					_, isPending, err := sendClient.Client.TransactionByHash(context.Background(), common.HexToHash(ord.SendTxHash))
					if err == nil && !isPending { // 查询到交易
						// 链上查询到交易，执行跨链转账中间地址发送交易
						// 跨链转账中间地址处理跨链转账后部分事务
						if err := process.ProcessCollectionTx(ord.FromAddr, ord.Amount); err != nil {
							log.Errorf("ethToTtProcess.ProcessCollectionTx(from, amount) error: %v", err)
						}
						return
					}

					if count == 0 {
						// 设置订单为失败，并退出
						_ = ord.Update(models.TacOrder{State: 2})
						return
					}
					time.Sleep(10 * time.Second)
				}
			}()

			continue
		} else {
			// 3. 最后情况为链上查询到交易之后的几种情况分析
			// 3.1 ReceiveTxHash不存在的情况
			if ord.ReceiveTxHash == "" {
				// 执行中间地址发送交易部分
				// 跨链转账中间地址处理跨链转账后部分事务
				if err := process.ProcessCollectionTx(ord.FromAddr, ord.Amount); err != nil {
					log.Errorf("ethToTtProcess.ProcessCollectionTx(from, amount) error: %v", err)
				}
				return

			} else {
				// 3.2 存在receiveTxHash
				// 链上查找是否上链
				_, _, err = receiveClient.Client.TransactionByHash(context.Background(), common.HexToHash(ord.ReceiveTxHash))
				if err == ethereum.NotFound {
					// 表示交易发送失败则中转地址需要重新发送
					// 跨链转账中间地址处理跨链转账后部分事务
					if err := process.ProcessCollectionTx(ord.FromAddr, ord.Amount); err != nil {
						log.Errorf("ethToTtProcess.ProcessCollectionTx(from, amount) error: %v", err)
					}
					return

				} else {
					// 设置订单状态为成功状态 todo 只要中间地址成功发送交易上链则代表一定会转账成功
					_ = ord.Update(models.TacOrder{State: 1})
					continue
				}
			}
		}
	}
}

func InitFlashOrderState() {

}
