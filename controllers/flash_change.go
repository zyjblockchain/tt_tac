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
		// 单位换算
		log.Infof("前端传入的闪兑参数；闪兑的usdt数量：%s; 接收的pala数量： %s", logic.FromTokenAmount, logic.ToTokenAmount)
		logic.FromTokenAmount = utils.FormatTokenAmount(logic.FromTokenAmount, 6) // usdt的单位换算
		logic.ToTokenAmount = utils.FormatTokenAmount(logic.ToTokenAmount, 8)     // pala单位换算
		// logic
		// 目前只有以太坊上的usdt闪兑pala，所以这里为了防止前端不传或者传错,重写token address
		logic.FromTokenAddress = conf.EthUSDTTokenAddress
		logic.ToTokenAddress = conf.EthPalaTokenAddress
		orderId, err := logic.FlashChange()
		if err != nil {
			if err == utils.VerifyPasswordErr {
				serializer.ErrorResponse(c, utils.CheckPasswordErrCode, utils.CheckPasswordErrMsg, "密码错误")
				return
			}
			log.Errorf("FlashChange logic err: %v", err)
			serializer.ErrorResponse(c, utils.SendFlashChangeTxErrCode, utils.SendFlashChangeTxErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, res{OrderId: orderId}, "success")
		}
	}
}

type ResParams struct {
	Page    uint   `json:"page"`
	Limit   uint   `json:"limit"`
	Address string `json:"address" binding:"required"`
}

type RespOrder struct {
	CreatedAt     int64  `json:"created_at"`
	ToTokenAmount string `json:"amount"` // pala token amount
	State         int    `json:"state"`  // 0. pending，1. success 2. failed 3. timeout

}
type RespResult struct {
	Total int         `json:"total"` // 总记录数
	List  []RespOrder `json:"list"`
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
		if resParams.Limit == 0 {
			resParams.Limit = 5
		}
		orders, total, err := new(models.FlashChangeOrder).GetBatchFlashOrder(resParams.Address, resParams.Page, resParams.Limit)
		if err != nil {
			log.Errorf("flashChange order get batch error: %v", err)
			serializer.ErrorResponse(c, utils.ExchangeOrderGetBatchErrCode, utils.ExchangeOrderGetBatchErrMsg, err.Error())
			return
		} else {
			resp := make([]RespOrder, 0, 0)
			for _, order := range orders {
				r := RespOrder{
					CreatedAt:     order.CreatedAt.Unix(),
					ToTokenAmount: utils.UnitConversion(order.ToTokenAmount, 8, 6),
					State:         order.State,
				}
				resp = append(resp, r)
			}

			serializer.SuccessResponse(c, RespResult{Total: total, List: resp}, "success")
		}
	}
}
