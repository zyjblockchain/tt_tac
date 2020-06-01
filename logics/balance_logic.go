package logics

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/utils"
	"github.com/zyjblockchain/tt_tac/utils/ding_robot"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"time"
)

type RespBalance struct {
	TtBalance  string `json:"tt_balance"`
	EthBalance string `json:"eth_balance"`
	Decimal    int    `json:"decimal"`
}
type GetBalance struct {
	Address string `json:"address" binding:"required"`
}

// GetBalance 获取主网余额
func (g *GetBalance) GetBalance() (*RespBalance, error) {
	ttClient := transaction.NewChainClient(conf.TTChainNet, big.NewInt(conf.TTChainID))
	defer ttClient.Close()
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()

	ttBalance, err := ttClient.Client.BalanceAt(context.Background(), common.HexToAddress(g.Address), nil)
	if err != nil {
		log.Errorf("获取tt链上的tt币 balance error: %v, address: %s", err, g.Address)
		return nil, err
	}

	EthBalance, err := ethClient.Client.BalanceAt(context.Background(), common.HexToAddress(g.Address), nil)
	if err != nil {
		log.Errorf("获取eth链上的eth币 balance error: %v, address: %s", err, g.Address)
		return nil, err
	}
	return &RespBalance{
		TtBalance:  utils.UnitConversion(ttBalance.String(), 18, 6),
		EthBalance: utils.UnitConversion(EthBalance.String(), 18, 6),
		Decimal:    18, // 这两个币的小数位数都是18位
	}, nil
}

type RespTokenBalance struct {
	TtPalaBalance  string `json:"tt_pala_balance"`
	EthPalaBalance string `json:"eth_pala_balance"`
	EthUsdtBalance string `json:"eth_usdt_balance"`
	UsdtDecimal    int    `json:"usdt_decimal"` // 6位小数
	PalaDecimal    int    `json:"pala_decimal"` // 8位
}
type TokenBalance struct {
	Address string `json:"address" binding:"required"`
}

// GetTokenBalance
func (t *TokenBalance) GetTokenBalance() (*RespTokenBalance, error) {
	ttClient := transaction.NewChainClient(conf.TTChainNet, big.NewInt(conf.TTChainID))
	defer ttClient.Close()
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	defer ethClient.Close()

	ttPalaBalance, err := ttClient.GetTokenBalance(common.HexToAddress(t.Address), common.HexToAddress(conf.TtPalaTokenAddress))
	if err != nil {
		log.Errorf("获取tt pala balance err:%v. address: %s", err, t.Address)
		return nil, err
	}

	EthPalaBalance, err := ethClient.GetTokenBalance(common.HexToAddress(t.Address), common.HexToAddress(conf.EthPalaTokenAddress))
	if err != nil {
		log.Errorf("获取eth pala balance err:%v. address: %s", err, t.Address)
		return nil, err
	}

	EthUsdtBalance, err := ethClient.GetTokenBalance(common.HexToAddress(t.Address), common.HexToAddress(conf.EthUSDTTokenAddress))
	if err != nil {
		log.Errorf("获取eth usdt balance err:%v. address: %s", err, t.Address)
		return nil, err
	}

	return &RespTokenBalance{
		TtPalaBalance:  utils.UnitConversion(ttPalaBalance.String(), 8, 6),
		EthPalaBalance: utils.UnitConversion(EthPalaBalance.String(), 8, 6),
		EthUsdtBalance: utils.UnitConversion(EthUsdtBalance.String(), 6, 6),
		UsdtDecimal:    6,
		PalaDecimal:    8,
	}, nil
}

type GetGasFee struct {
	ChainTag int `json:"chain_tag" binding:"required"`
}

type Fee struct {
	GasFee string `json:"gas_fee"`
}

func (g *GetGasFee) GetGasFee() (*Fee, error) {
	var chainUrl string
	var chainId *big.Int
	switch g.ChainTag {
	case conf.EthChainTag:
		chainUrl = conf.EthChainNet
		chainId = big.NewInt(conf.EthChainID)
	case conf.TTChainTag:
		chainUrl = conf.TTChainNet
		chainId = big.NewInt(conf.TTChainID)
	default:
		return nil, errors.New(fmt.Sprintf("输入的chain_tag不存在。tt chain_tag = %d; ethereum chain_tag = %d", conf.TTChainTag, conf.EthChainTag))
	}
	client := transaction.NewChainClient(chainUrl, chainId)
	defer client.Close()
	suggestPrice, err := client.SuggestGasPrice()
	if err != nil {
		log.Errorf("get suggest gas price err: %v", err)
		return nil, err
	}
	gasPrice := new(big.Int).Mul(suggestPrice, big.NewInt(2)) // 两倍于gasPrice
	log.Infof("gasPrice: %s", gasPrice.String())
	gasLimit := uint64(60000)
	gasFee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit))).String()
	return &Fee{GasFee: utils.UnitConversion(gasFee, 18, 6)}, nil
}

// CheckMiddleAddressBalance 定时检查中间地址的各种资产的balance是否足够
func CheckMiddleAddressBalance() {
	dingRobot := ding_robot.NewRobot(conf.BalanceWebHook)
	log.Infof("中转地址balance钉钉告警webHook: %s", conf.BalanceWebHook)
	getBalanceTicker := time.NewTicker(30 * time.Second)
	ethClient := transaction.NewChainClient(conf.EthChainNet, big.NewInt(int64(conf.EthChainID)))
	ttClient := transaction.NewChainClient(conf.TTChainNet, big.NewInt(int64(conf.TTChainID)))
	defer func() {
		ethClient.Close()
		ttClient.Close()
	}()
	for {
		select {
		case <-getBalanceTicker.C:
			// 1. 查询tac中间地址的eth余额
			getTacMiddleEthBalance, err := ethClient.Client.BalanceAt(context.Background(), common.HexToAddress(conf.TacMiddleAddress), nil)
			if err != nil {
				log.Errorf("查询tac中间地址的eth余额 error: %v", err)
			} else {
				// 最小余额限度0.5 eth
				limitBalance, _ := new(big.Int).SetString("500000000000000000", 10)
				if getTacMiddleEthBalance.Cmp(limitBalance) < 0 {
					// 通知需要充eth了
					content := fmt.Sprintf("1.跨链转账中转地址eth余额即将消耗完;\naddress: %s,\nbalance: %s eth", conf.TacMiddleAddress, utils.UnitConversion(getTacMiddleEthBalance.String(), 18, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}

			// 2. 查询tac中间地址的eth上的pala
			getTacMiddleEthPalaBalance, err := ethClient.GetTokenBalance(common.HexToAddress(conf.TacMiddleAddress), common.HexToAddress(conf.EthPalaTokenAddress))
			if err != nil {
				log.Errorf("查询tac中间地址的以太坊上的pala余额 error: %v", err)
			} else {
				// 最小余额限度 1000 pala
				limitBalance, _ := new(big.Int).SetString("100000000000", 10)
				if getTacMiddleEthPalaBalance.Cmp(limitBalance) < 0 {
					content := fmt.Sprintf("2.跨链转账中转地址以太坊上的pala余额即将消耗完;\naddress: %s,\nbalance:  %s eth", conf.TacMiddleAddress, utils.UnitConversion(getTacMiddleEthPalaBalance.String(), 8, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}

			// 3. 查询闪兑中间地址的eth余额
			getFlashMiddleEthBalance, err := ethClient.Client.BalanceAt(context.Background(), common.HexToAddress(conf.EthFlashChangeMiddleAddress), nil)
			if err != nil {
				log.Errorf("查询闪兑中间地址的eth余额 error: %v", err)
			} else {
				// 最小余额限度0.5 eth
				limitBalance, _ := new(big.Int).SetString("500000000000000000", 10)
				if getFlashMiddleEthBalance.Cmp(limitBalance) < 0 {
					// 通知需要充eth了
					content := fmt.Sprintf("3.闪兑中转地址eth余额即将消耗完;\naddress: %s,\nbalance: %s eth", conf.EthFlashChangeMiddleAddress, utils.UnitConversion(getFlashMiddleEthBalance.String(), 18, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}

			// 4. 查询闪兑中间地址的eth上的pala余额
			getFlashMiddleEthPalaBalance, err := ethClient.GetTokenBalance(common.HexToAddress(conf.EthFlashChangeMiddleAddress), common.HexToAddress(conf.EthPalaTokenAddress))
			if err != nil {
				log.Errorf("查询闪兑中间地址的以太坊上的pala余额 error: %v", err)
			} else {
				// 最小余额限度 1000 pala
				limitBalance, _ := new(big.Int).SetString("100000000000", 10)
				if getFlashMiddleEthPalaBalance.Cmp(limitBalance) < 0 {
					content := fmt.Sprintf("4.闪兑中转地址以太坊上的pala余额即将消耗完;\naddress: %s,\neth_pala_balance: %s eth", conf.EthFlashChangeMiddleAddress, utils.UnitConversion(getFlashMiddleEthPalaBalance.String(), 8, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}

			// 5. 查询跨链转账中间地址的tt余额
			getTacMiddleTTBalance, err := ttClient.Client.BalanceAt(context.Background(), common.HexToAddress(conf.TacMiddleAddress), nil)
			if err != nil {
				log.Errorf("查询跨链转账中间地址的tt余额 error: %v", err)
			} else {
				// 最小余额限度0.5 tt
				limitBalance, _ := new(big.Int).SetString("500000000000000000", 10)
				if getTacMiddleTTBalance.Cmp(limitBalance) < 0 {
					// 通知需要充tt了
					content := fmt.Sprintf("5.查询跨链转账中间地址的tt余额即将消耗完;\naddress: %s,\ntt_balance: %s tt", conf.TacMiddleAddress, utils.UnitConversion(getTacMiddleTTBalance.String(), 18, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}

			// 6. 查询跨链转账中间地址的tt链上的pala余额
			getTacMiddleTTPalaBalance, err := ttClient.GetTokenBalance(common.HexToAddress(conf.TacMiddleAddress), common.HexToAddress(conf.TtPalaTokenAddress))
			if err != nil {
				log.Errorf("查询跨链转账中间地址的tt链上的pala余额 error: %v", err)
			} else {
				// 最小余额限度 1000 pala
				limitBalance, _ := new(big.Int).SetString("100000000000", 10)
				if getTacMiddleTTPalaBalance.Cmp(limitBalance) < 0 {
					// 通知需要充tt了
					content := fmt.Sprintf("6.查询跨链转账中间地址的tt链上的pala余额即将消耗完;\naddress: %s,\ntt_pala_balance: %s eth", conf.TacMiddleAddress, utils.UnitConversion(getTacMiddleTTPalaBalance.String(), 18, 6))
					_ = dingRobot.SendText(content, nil, true)
				}
			}
		}
	}
}
