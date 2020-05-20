package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
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
