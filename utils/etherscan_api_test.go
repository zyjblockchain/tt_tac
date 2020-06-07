package utils

import (
	"testing"
)

func TestEtherscanApi(t *testing.T) {
	contractAddress := "0xD20fb5cf926Dc29c88f64725e6f911f40f7bf531"
	address := "0xF9891E1A2635CB8D8C25A6A2ec8E453bFb2E67c4"
	page := 1
	offset := 5
	txs, err := GetAddressTokenTransfers(contractAddress, address, page, offset)
	t.Log(err)
	for _, tx := range txs {
		t.Log(tx.Hash)
	}

}

func TestGetAddressTokenTransfers(t *testing.T) {
	txs, err := GetAddressEthTransfers("0xb378413ef8b086628d1f0f01fef785ab501970fa", 0, 0)
	t.Log(err)
	for _, tx := range txs {
		t.Log(tx.Hash)
	}
}
