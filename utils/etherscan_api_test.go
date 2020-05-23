package utils

import (
	"testing"
	"time"
)

func TestEtherscanApi(t *testing.T) {
	contractAddress := "0x3f69636Af46718cBd27002c65256226742309E1f"
	address := "0xb378413ef8b086628d1f0f01fef785ab501970fa"
	page := 1
	offset := 5
	txs, err := GetAddressTokenTransfers(contractAddress, address, page, offset)
	t.Log(err)
	for _, tx := range txs {
		t.Log(tx.Hash)
	}

}

func TestGetAddressTokenTransfers(t *testing.T) {
	tt := time.Now().Unix()
	t.Log(tt)
}
