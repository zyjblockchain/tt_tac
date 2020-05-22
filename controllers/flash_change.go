package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/models"
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

type ResParams struct {
	StartIndex uint   `json:"start_index"`
	Limit      uint   `json:"limit" binding:"required"`
	Address    string `json:"address" binding:"required"`
}

type RespResult struct {
	CreatedAt     int64  `json:"created_at"`
	ToTokenAmount string `json:"amount"` // pala token amount
	State         int    `json:"state"`  // 0. pending，1. success 2. failed 3. timeout
}

// GetBatchOrderByAddress 分页拉取地址的闪兑记录
func GetBatchOrderByAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		var resParams ResParams
		err := c.ShouldBind(&resParams)
		if err != nil {
			log.Errorf("FlashChange should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// 拉取地址下的闪兑订单
		orders, err := new(models.FlashChangeOrder).GetBatchFlashOrder(resParams.Address, resParams.StartIndex, resParams.Limit)
		if err != nil {
			log.Errorf("flashChange order get batch error: %v", err)
			serializer.ErrorResponse(c, utils.ExchangeOrderGetBatchErrCode, utils.ExchangeOrderGetBatchErrMsg, err.Error())
			return
		} else {
			var resp []RespResult
			for _, order := range orders {
				r := RespResult{
					CreatedAt:     order.CreatedAt.Unix(),
					ToTokenAmount: order.ToTokenAmount,
					State:         order.State,
				}
				resp = append(resp, r)
			}
			serializer.SuccessResponse(c, resp, "success")
		}
	}
}
