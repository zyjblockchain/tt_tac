package models

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zyjblockchain/tt_tac/conf"
	transaction "github.com/zyjblockchain/tt_tac/utils/tx_utils"
	"math/big"
	"testing"
)

func TestGetKv(t *testing.T) {
	dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	InitDB(dsn)
	txHash := "0xef4f851b881beabd83ab95e20dca849bd18f60630d7afdd0ee54e0c93ec6646a"

	hexTx, err := GetKv(txHash)
	t.Log(err)
	t.Log(hexTx)

	tx := &types.Transaction{}
	err = tx.UnmarshalJSON(hexTx)
	t.Log(err)
	t.Log(tx.Hash().String(), tx.ChainId(), tx.GasPrice().String())
}

func TestSetKv(t *testing.T) {
	dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	InitDB(dsn)

	client := transaction.NewChainClient(conf.EthChainNet, big.NewInt(conf.EthChainID))
	fromPrivate := "69F657EAF364969CCFB2531F25D9C9EFAC0A631159CEA51E5F7D834078411111"
	nonce := uint64(32)
	gasLimit := uint64(60000)
	gasPrice := big.NewInt(200000000000)
	to := common.HexToAddress("0x7c3d8e14b56a3229164e3dbe1536336d940fae32ab641e97f5e900797474552d")
	tokenAddress := common.HexToAddress("0xeff6f1612d03205BA5E8d26cAc1397bf778ab1AC")
	tokenAmount := big.NewInt(8888888)
	signedTx, err := client.NewSignedTokenTx(fromPrivate, nonce, gasLimit, gasPrice, to, tokenAddress, tokenAmount)
	t.Log(err)
	byteTx, err := signedTx.MarshalJSON()
	t.Log(byteTx)
	t.Log(len(byteTx))
	t.Log(err)

	_ = SetKv(signedTx.Hash().String(), byteTx)

	tx := &types.Transaction{}
	err = tx.UnmarshalJSON(byteTx)
	t.Log(err)
	t.Log(tx.Hash().String(), tx.ChainId())

	txHash := "0xef4f851b881beabd83ab95e20dca849bd18f60630d7afdd0ee54e0c93ec6646a"

	hexTx, err := GetKv(txHash)
	t.Log(err)
	t.Log(len(hexTx))
	tt := &types.Transaction{}
	err = json.Unmarshal(hexTx, tt)
	t.Log(err)
}
