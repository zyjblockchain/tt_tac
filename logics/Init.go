package logics

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/models"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"strings"
	"time"
)

// InitTacOrderState 服务重启之后初始化跨链转账的订单状态
func InitTacOrderState(ethToTtProcess, ttToEthProcess *TacProcess) {
	ttClient := transaction.NewChainClient(conf.TTChainNet, big.NewInt(conf.TTChainID))
	defer ttClient.Close()
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()
	// 获取tacOrder表中状态为0的记录
	tacOrds, err := new(models.TacOrder).GetTacOrdersByState(0)
	if err != nil {
		log.Errorf("1. 获取tacOrder表中状态为0的记录失败；error: %v", err)
		return
	}
	log.Infof("2. 开始遍历查询出来的tac order。订单数量：%d", len(tacOrds))
	for _, ord := range tacOrds {
		// 设置process
		process := ethToTtProcess // 默认值
		if ord.OrderType == conf.EthToTtOrderType {

		} else if ord.OrderType == conf.TtToEthOrderType {
			process = ttToEthProcess
		} else {
			log.Errorf("2.1 ord.OrderType 不存在则删除此跨链转账记录")
			// 错误的订单，直接删除
			_ = ord.Delete(ord.ID)
			continue
		}

		// 1. 如果没有send_Tx_hash，则说明跨链转账申请者转账到中间地址的交易都失败了
		if ord.SendTxHash == "" {
			log.Infof("2.2 tacOrder中不存在SendTxHash则删除")
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
			log.Errorf("2.3 该跨链转账订单的orderType不正确。orderId: %d, orderType: %d", ord.ID, ord.OrderType)
			_ = ord.Delete(ord.ID)
			continue
		}
		// 2.2 通过交易hash查询交易是否上链发送成功
		_, isPending, err := sendClient.Client.TransactionByHash(context.Background(), common.HexToHash(ord.SendTxHash))
		if err != nil { // todo 此处应该判断err ==  ethereum.NotFound,但是我们默认为我们发送的交易只要执行都会被执行成功的，也有可能网络错误，但是这里我们不考虑这种情况了
			// 交易没有发送成功，则delete记录
			log.Errorf("3.1 跨链转账申请者send交易没有上链，删除订单")
			_ = ord.Delete(ord.ID)
			continue
		} else if isPending { // 这种情况可能性很少
			log.Infof("3.2 跨链转账申请者发送的交易在pending中，txHash: %s", ord.SendTxHash)
			// 2.3 此交易正在pending
			// 开启协程轮询监听交易情况
			go func() {
				log.Infof("3.3 跨链转账开始轮询申请者的send交易： %s", ord.SendTxHash)
				var count = 5
				for {
					count--
					time.Sleep(10 * time.Second)
					_, isPending, err := sendClient.Client.TransactionByHash(context.Background(), common.HexToHash(ord.SendTxHash))
					if err == nil && !isPending { // 查询到交易
						// 链上查询到交易，执行跨链转账中间地址发送交易
						// 跨链转账中间地址处理跨链转账后部分事务
						log.Infof("3.4 跨链转账轮询到申请者的send交易，开始执行ProcessCollectionTx; hash: %s", ord.SendTxHash)
						if err := process.ProcessCollectionTx(ord.FromAddr, ord.Amount); err != nil {
							log.Errorf("ethToTtProcess.ProcessCollectionTx(from, amount) error: %v", err)
						}
						return
					}

					if count == 0 {
						log.Errorf("3.5 跨链转账轮询申请者的send交易超时，订单删除; hash: %s", ord.SendTxHash)
						// 设置订单为失败，并退出
						_ = ord.Update(models.TacOrder{State: 2})
						return
					}
				}
			}()

			continue
		} else {
			log.Infof("4.1 跨链转账链上查询到申请者的send交易； hash: %s", ord.SendTxHash)
			// 3. 最后情况为链上查询到交易之后的几种情况分析
			// 3.1 ReceiveTxHash不存在的情况
			if ord.ReceiveTxHash == "" {
				log.Infof("4.2 跨链转账中没有中间地址的转账交易hash,则开始执行ProcessCollectionTx函数，完成跨链转账的后半部分")
				// 执行中间地址发送交易部分
				// 跨链转账中间地址处理跨链转账后部分事务
				if err := process.ProcessCollectionTx(ord.FromAddr, ord.Amount); err != nil {
					log.Errorf("ethToTtProcess.ProcessCollectionTx(from, amount) error: %v", err)
				}
				continue

			} else {
				log.Infof("4.3 跨链转账存在ReceiveTxHash，检查交易是否上链。hash: %s", ord.ReceiveTxHash)
				// 3.2 存在receiveTxHash
				// 链上查找是否上链
				_, _, err = receiveClient.Client.TransactionByHash(context.Background(), common.HexToHash(ord.ReceiveTxHash))
				if err == ethereum.NotFound {
					log.Errorf("4.4 跨链转账查询ord.ReceiveTxHash 链上不存在，则开启ProcessCollectionTx 跨链转账后部分; hash: %s", ord.ReceiveTxHash)
					// 表示交易发送失败则中转地址需要重新发送
					// 跨链转账中间地址处理跨链转账后部分事务
					if err := process.ProcessCollectionTx(ord.FromAddr, ord.Amount); err != nil {
						log.Errorf("ethToTtProcess.ProcessCollectionTx(from, amount) error: %v", err)
					}
					continue

				} else {
					log.Infof("4.5 跨链转账订单满足完成情况，则置位成功状态，tacOrderId: %d", ord.ID)
					// 设置订单状态为成功状态 todo 只要中间地址成功发送交易上链则代表一定会转账成功
					_ = ord.Update(models.TacOrder{State: 1})
					continue
				}
			}
		}
	}
}

// InitTacOrderState 服务重启之后闪兑的订单状态
func InitFlashOrderState(flashSvr *WatchFlashChange) {
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()
	// 获取闪兑订单中状态为pending的记录
	flashOrds, err := new(models.FlashChangeOrder).GetFlashOrdersByState(0)
	if err != nil {
		log.Errorf("获取FlashChangeOrder表中状态为0的记录失败；error: %v", err)
		return
	}
	log.Infof("开始遍历查询出来的flash change order。订单数量：%d", len(flashOrds))
	for _, ord := range flashOrds {
		// 记录到缓存中
		FlashAddressMap[strings.ToLower(ord.OperateAddress)] = 1
		// 1. 查看是否有SendTxId
		if ord.SendTxId == 0 {
			log.Infof("1. 闪兑订单中没有sendTxId,则删除订单")
			// 直接删除订单
			_ = ord.Delete(ord.ID)
			continue
		}
		// 2. 存在则查看是否上链成功
		txTr, err := (&models.TxTransfer{Model: gorm.Model{ID: ord.SendTxId}}).Get()
		if err != nil {
			log.Errorf("2. 通过闪兑订单中的sendTxId查询txTransfer失败")
			// 直接删除订单
			_ = ord.Delete(ord.ID)
			continue
		}
		sendTxHash := txTr.TxHash
		log.Infof("2.2 闪兑sendTxHash: %s", sendTxHash)
		_, isPending, err := ethClient.Client.TransactionByHash(context.Background(), common.HexToHash(sendTxHash))
		if err != nil {
			log.Errorf("2.3 闪兑申请者send的usdt交易没有上链, err: %v", err)
			_ = ord.Delete(ord.ID)
			continue
		} else if isPending {
			// 2.1 此交易正在pending
			log.Infof("2.4 闪兑申请者发送的交易正在pending,开启交易监听; txHash: %s", sendTxHash)
			// 开启协程轮询监听交易情况
			go func() {
				var count = 5
				for {
					count--
					time.Sleep(20 * time.Second)
					_, isPending, err := ethClient.Client.TransactionByHash(context.Background(), common.HexToHash(sendTxHash))
					if err == nil && !isPending { // 查询到交易
						log.Infof("2.5 监听到闪兑申请者发送的usdt交易，开启发送pala给申请者; txHash: %s", sendTxHash)
						// 则执行闪兑中间地址转账部分
						if err := flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount); err != nil {
							log.Errorf("flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount) error: %v", err)
						}
						return
					}

					if count == 0 {
						// 设置订单为失败，并退出
						log.Errorf("监听申请者的闪兑发送的usdt交易失败，闪兑订单置位失败状态。txHash: %s", sendTxHash)
						_ = ord.Update(models.FlashChangeOrder{State: 2})
						return
					}
				}
			}()

			continue
		} else {
			log.Infof("3.0 闪兑sendTxHash链上查询到了；hash: %s", sendTxHash)
			// 3. 查询到了交易，则查看中间地址是否转了pala完成了闪兑的后半部分
			// 3.1 中间地址没有开始转 pala交易， 则执行后半部分事务
			if ord.ReceiveTxId == 0 {
				// 则执行闪兑中间地址转账部分
				log.Infof("3.1 闪兑中间地址发送pala给申请者账户；address: %s, usdtAmount: %s", ord.OperateAddress, ord.FromTokenAmount)
				if err := flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount); err != nil {
					log.Errorf("flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount) error: %v", err)
				}
				continue
			} else {
				// 3.2 存在则查看交易详情
				txTr, err := (&models.TxTransfer{Model: gorm.Model{ID: ord.ReceiveTxId}}).Get()
				if err != nil {
					log.Errorf("3.2 通过闪兑订单中的ReceiveTxId查询txTransfer失败")
					// 直接删除订单
					_ = ord.Delete(ord.ID)
					continue
				}
				receiveTxHash := txTr.TxHash
				_, _, err = ethClient.Client.TransactionByHash(context.Background(), common.HexToHash(receiveTxHash))
				if err == ethereum.NotFound {
					log.Infof("3.3 闪兑 receiveTxHash：%s 在链上查询不到，则开启中间地址发送pala给申请者账户address: %s, usdtAmount: %s", receiveTxHash, ord.OperateAddress, ord.FromTokenAmount)
					// 表示交易发送失败则中转地址需要重新发送
					if err := flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount); err != nil {
						log.Errorf("flashSvr.ProcessCollectFlashChangeTx(ord.OperateAddress, ord.FromTokenAmount) error: %v", err)
					}
					continue
				} else {
					log.Infof("3.4 闪兑订单满足完成条件，把订单置位成功状态；flashOrderId: %d", ord.ID)
					// 设置订单状态为成功状态 todo 只要中间地址成功发送交易上链则代表一定会转账成功
					_ = ord.Update(models.FlashChangeOrder{State: 1})
					continue
				}
			}
		}
	}
}
