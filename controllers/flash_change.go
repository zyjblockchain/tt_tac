package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
)

// FlashChange 闪兑
func FlashChange() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.FlashChange
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("FlashChange should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		// 目前只有以太坊上的usdt闪兑pala，所以这里为了防止前端不传或者传错,重写token address
		logic.FromTokenAddress = conf.EthUSDTTokenAddress
		logic.ToTokenAddress = conf.EthPalaTokenAddress
		orderId, err := logic.FlashChange()
		if err != nil {
			log.Errorf("FlashChange logic err: %v", err)
			serializer.ErrorResponse(c, utils.SendFlashChangeTxErrCode, utils.SendFlashChangeTxErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, res{OrderId: orderId}, "success")
		}
	}
}
