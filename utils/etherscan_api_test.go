package utils

import (
	"testing"
	"time"
)

func TestEtherscanApi(t *testing.T) {
	contractAddress := "0x3f69636af46718cbd27002c65256226742309e1f"
	address := "0x1EA6cef67DCc6A0a471b1AC2BaFb3e85dc6C6e18"
	page := 1
	offset := 1
	for i := 0; i < 10; i++ {
		txs, err := GetAddressTokenTransfers(contractAddress, address, page, offset)
		t.Log(err)
		t.Log(txs[0].Hash)
	}

}

func TestGetAddressTokenTransfers(t *testing.T) {
	tt := time.Now().Unix()
	t.Log(tt)
}
