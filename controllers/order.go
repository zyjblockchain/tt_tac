package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
)

type res struct {
	OrderId uint `json:"orderId"`
}

func ApplyOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var svr logics.Order
		err := c.ShouldBind(&svr)
		if err != nil {
			log.Errorf("ApplyOrder should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
		}
		// logic
		orderId, err := svr.CreateOrder()
		if err != nil {
			log.Errorf("create order logic err: %v", err)
			serializer.ErrorResponse(c, utils.OrderLogicErrCode, utils.OrderLogicErrMsg, err.Error())
		} else {
			serializer.SuccessResponse(c, res{OrderId: orderId}, "success")
		}
	}
}
