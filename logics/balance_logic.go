package logics

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
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
		TtBalance:  ttBalance.String(),
		EthBalance: EthBalance.String(),
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
		TtPalaBalance:  ttPalaBalance.String(),
		EthPalaBalance: EthPalaBalance.String(),
		EthUsdtBalance: EthUsdtBalance.String(),
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
	return &Fee{GasFee: gasFee}, nil
}
