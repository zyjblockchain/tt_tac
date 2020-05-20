package btc_max_api

import "testing"

func TestGetSingleMarketTicker(t *testing.T) {
	tt, err := GetSingleMarketTicker("PALA_USDT")
	t.Log(err)
	t.Log(*tt)
}
