package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
	"github.com/zyjblockchain/tt_tac/utils/btc_max_api"
)

// GetBalance 获取tt链和eth链的address对应的主网币的余额
func GetBalance() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.GetBalance
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("GetBalance should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		respBalance, err := logic.GetBalance()
		if err != nil {
			log.Errorf("GetBalance logic err: %v", err)
			serializer.ErrorResponse(c, utils.GetBalanceErrCode, utils.GetBalanceErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, *respBalance, "success")
		}
	}
}

// GetTokenBalance
func GetTokenBalance() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.TokenBalance
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("GetTokenBalance should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		respTokenBalance, err := logic.GetTokenBalance()
		if err != nil {
			log.Errorf("GetTokenBalance logic err: %v", err)
			serializer.ErrorResponse(c, utils.GetTokenBalanceErrCode, utils.GetTokenBalanceErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, *respTokenBalance, "success")
		}
	}
}

// GetLatestPalaToUsdtPrice
func GetLatestPalaToUsdtPrice() gin.HandlerFunc {
	return func(c *gin.Context) {
		pair := "PALA_USDT"
		var err error
		var tt *btc_max_api.Ticker
		tt, err = btc_max_api.GetSingleMarketTicker(pair)
		if err != nil {
			// 重新请求一次
			tt, err = btc_max_api.GetSingleMarketTicker(pair)
		}
		if err != nil {
			// 返回error给前端
			log.Errorf("GetSingleMarketTicker should binding error: %v", err)
			serializer.ErrorResponse(c, utils.GetLatestPriceErrCode, utils.GetLatestPriceErrMsg, err.Error())
			return
		} else {
			// 返回结果
			// 给前端展示的pala价格需要高于真实价格
			inc := decimal.NewFromFloat(conf.FlashPalaToUsdtPriceChange) // 配置文件中默认上浮1%
			price, err := decimal.NewFromString(tt.TradePrice)
			if err != nil {
				log.Errorf(" decimal.NewFromString(tt.TradePrice)  error: %v", err)
				serializer.ErrorResponse(c, utils.GetLatestPriceErrCode, utils.GetLatestPriceErrMsg, err.Error())
				return
			}
			tt.TradePrice = price.Mul(inc).String()
			serializer.SuccessResponse(c, *tt, "success")
		}
	}
}

// GetLatestEthToUsdtPrice
func GetLatestEthToUsdtPrice() gin.HandlerFunc {
	return func(c *gin.Context) {
		pair := "ETH_USDT"
		var err error
		var tt *btc_max_api.Ticker
		tt, err = btc_max_api.GetSingleMarketTicker(pair)
		if err != nil {
			// 重新请求一次
			tt, err = btc_max_api.GetSingleMarketTicker(pair)
		}
		if err != nil {
			// 返回error给前端
			log.Errorf("GetSingleMarketTicker should binding error: %v", err)
			serializer.ErrorResponse(c, utils.GetLatestPriceErrCode, utils.GetLatestPriceErrMsg, err.Error())
			return
		} else {
			// 返回结果
			serializer.SuccessResponse(c, *tt, "success")
		}
	}
}

// 拉取tx gas fee
func GetGasFee() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.GetGasFee
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("GetGasFee should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		fee, err := logic.GetGasFee()
		if err != nil {
			log.Errorf("GetSingleMarketTicker should binding error: %v", err)
			serializer.ErrorResponse(c, utils.GetGasFeeErrCode, utils.GetGasFeeErrMsg, err.Error())
			return
		} else {
			log.Infof("获取chainTag: %d 的一笔交易的建议gas fee: %s", logic.ChainTag, fee.GasFee)
			serializer.SuccessResponse(c, *fee, "success")
		}
	}
}
