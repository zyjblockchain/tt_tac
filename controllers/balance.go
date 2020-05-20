package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
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