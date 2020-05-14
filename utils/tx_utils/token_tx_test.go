package transaction

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/zyjblockchain/tt_tac/utils"
	"math/big"
	"testing"
)

const (
	TTMainNet  = "https://mainnet-rpc.thundercore.com"
	EthTestNet = "https://rinkeby.infura.io/v3/36b98a13557c4b8583d57934ede2f74d"

	TTMainNetID  = 108
	EthTestNetID = 4
)

func TestGetTokenBalance(t *testing.T) {
	client := NewChainClient(TTMainNet, big.NewInt(TTMainNetID))
	defer client.Close()
	address := "0x67Adf250F70F6100d346cF8FE3af6DC7A2C23213"
	tokenAddr := "0x087cC4Aaa83aCA54bDCC89920483c8e2a30Bc47c"
	balance, err := client.GetTokenBalance(common.HexToAddress(address), common.HexToAddress(tokenAddr))
	t.Log(err)
	t.Log(balance.String())
}

func TestChainClient_EstimateTokenTxGas(t *testing.T) {
	client := NewChainClient(TTMainNet, big.NewInt(TTMainNetID))
	defer client.Close()

	tokenAmount := big.NewInt(900000000)
	from := "0x67Adf250F70F6100d346cF8FE3af6DC7A2C23213"
	tokenAddr := "0x087cC4Aaa83aCA54bDCC89920483c8e2a30Bc47c"
	receiver := "0x67Adf250F70F6100d346cF8FE3af6DC7A2C23213"
	gasLimit, err := client.EstimateTokenTxGas(tokenAmount, common.HexToAddress(from), common.HexToAddress(tokenAddr), common.HexToAddress(receiver))
	t.Log(err)
	t.Log(gasLimit)
}

func TestChainClient_GetNonce(t *testing.T) {
	client := NewChainClient(TTMainNet, big.NewInt(TTMainNetID))
	defer client.Close()
	nonce, err := client.GetNonce(common.HexToAddress("0x67Adf250F70F6100d346cF8FE3af6DC7A2C23213"))
	t.Log(err)
	t.Log(nonce)
}

func TestChainClient_SuggestGasPrice(t *testing.T) {
	client := NewChainClient(TTMainNet, big.NewInt(TTMainNetID))
	defer client.Close()
	gasPrice, err := client.SuggestGasPrice()
	t.Log(err)
	t.Log(gasPrice)
}

func TestChainClient_SendTokenTx(t *testing.T) {
	client := NewChainClient(TTMainNet, big.NewInt(TTMainNetID))
	defer client.Close()
	prv := "61086E09073DCCF0A03D9D1BE953E161532A264A959C0608158B6C9ACA92D25B"
	addr, _ := utils.PrivateToAddress(prv)
	nonce, _ := client.GetNonce(addr)

	tokenAddress := "0x087cC4Aaa83aCA54bDCC89920483c8e2a30Bc47c" // tt 主网上的sandy代币
	tokenAmount := big.NewInt(900000000)
	recieve := "0x59375A522876aB96B0ed2953D0D3b92674701Cc2"
	gasLimit, _ := client.EstimateTokenTxGas(tokenAmount, addr, common.HexToAddress(tokenAddress), common.HexToAddress(recieve))

	gasPrice, _ := client.SuggestGasPrice()
	tx, err := client.SendTokenTx(prv, nonce, gasLimit, gasPrice, common.HexToAddress(recieve), common.HexToAddress(tokenAddress), tokenAmount)
	t.Log(err)
	t.Log(tx.Hash().String())
}

func TestChainClient_Close(t *testing.T) {

	// amount, ok := new(big.Int).SetString("0", 10)
	// t.Log(ok)
	t.Log(big.NewInt(0).String() == "0")
}
