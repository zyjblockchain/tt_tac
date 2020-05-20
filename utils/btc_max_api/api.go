package btc_max_api

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/zyjblockchain/sandy_log/log"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	BTCMAX_API_DEMAIN = "https://api.btcmax.com"
)

type Ticker struct {
	Pair       string `json:"pair"`        // PALA_USDT
	TradePrice string `json:"trade_price"` // 最新的交易价格
}

// GetSingleMarketTicker pair = PALA_USDT 或者 ETH_USDT
func GetSingleMarketTicker(pair string) (*Ticker, error) {
	param := fmt.Sprintf("pair=%s", pair)
	url := BTCMAX_API_DEMAIN + "/openapi1/pair" + "?" + param
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("get 请求失败，error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	// 反序列化body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("read all err: %v", err)
		return nil, err
	}
	js, err := simplejson.NewJson(b)
	if err != nil {
		log.Errorf("simple json err: %v", err)
		return nil, err
	}

	price, err := js.Get("data").Get(strings.ToUpper(pair)).Get("ticker").Get("trade_price").String()
	if err != nil {
		log.Errorf("simple json 解析trade_price失败： %v", err)
		return nil, err
	}
	log.Infof("price: %v", price)

	tt := &Ticker{
		Pair:       strings.ToUpper(pair),
		TradePrice: price,
	}
	return tt, nil
}
