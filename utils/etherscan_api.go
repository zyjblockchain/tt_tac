package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nanmu42/etherscan-api"
	"github.com/zyjblockchain/sandy_log/log"
	"io"
	"math/big"
	"net/http"
)

var apiKeyPool = []string{"BBDQ1WWK1MYHTHTKDS8HGFW8CJPHKU6U26", "Q8ZXAU8MQGGDDDVAYZV351I9RYHK8VXNIK", "RHGF4ZTBRQ8BKND1WZ8RIYZY7W21T5B33I"}

// GetAddressTokenTransfers 拉取地址的token交易
func GetAddressTokenTransfers(contractaddress, address string, page, offset int) ([]etherscan.ERC20Transfer, error) {
	// 随机获取apiKeyPool中的apiKey
	var index = 0
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(apiKeyPool))))
	if err == nil {
		index = int(n.Int64())
	}
	apiKey := apiKeyPool[index]

	url := fmt.Sprintf("https://api-cn.etherscan.com/api?module=account&action=tokentx&contractaddress=%s&address=%s&page=%d&offset=%d&sort=desc&apikey=%s", contractaddress, address, page, offset, apiKey)
	// url := fmt.Sprintf("https://api-rinkeby.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&address=%s&page=%d&offset=%d&sort=desc&apikey=%s", contractaddress, address, page, offset, apiKey) // for test
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("get etherscan err: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var content bytes.Buffer
	if _, err := io.Copy(&content, resp.Body); err != nil {
		log.Errorf("io copy err : %v", err)
		return nil, err
	}

	var envelope etherscan.Envelope
	err = json.Unmarshal(content.Bytes(), &envelope)
	if err != nil {
		log.Errorf("unmarshal resp error: %v", err)
		return nil, err
	}
	var txs []etherscan.ERC20Transfer
	err = json.Unmarshal(bytes.Replace(envelope.Result, []byte(`"tokenDecimal":""`), []byte(`"tokenDecimal":"0"`), -1), &txs)
	if err != nil {
		log.Errorf("unmarshal resp error: %v", err)
		return nil, err
	}
	return txs, nil
}

// GetAddressEthTransfers 拉取地址的以太坊转账交易
func GetAddressEthTransfers(address string, page, offset int) ([]etherscan.NormalTx, error) {
	// 随机获取apiKeyPool中的apiKey
	var index = 0
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(apiKeyPool))))
	if err == nil {
		index = int(n.Int64())
	}
	apiKey := apiKeyPool[index]

	url := fmt.Sprintf("https://api-cn.etherscan.com/api?action=txlist&address=%s&apikey=%s&module=account&offset=%d&page=%d&sort=desc", address, apiKey, offset, page)
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("get etherscan err: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var content bytes.Buffer
	if _, err := io.Copy(&content, resp.Body); err != nil {
		log.Errorf("io copy err : %v", err)
		return nil, err
	}
	var envelope etherscan.Envelope
	err = json.Unmarshal(content.Bytes(), &envelope)
	if err != nil {
		log.Errorf("unmarshal resp error: %v", err)
		return nil, err
	}

	var txs []etherscan.NormalTx
	err = json.Unmarshal(envelope.Result, &txs)
	if err != nil {
		log.Errorf("unmarshal resp error: %v", err)
		return nil, err
	}
	return txs, nil
}
