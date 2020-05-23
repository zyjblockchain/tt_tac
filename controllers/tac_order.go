package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
	"strconv"
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
			return
		}
		// 单位换算
		svr.Amount = utils.FormatTokenAmount(svr.Amount, 8)
		// logic
		orderId, err := svr.CreateOrder()
		if err != nil {
			log.Errorf("create order logic err: %v", err)
			serializer.ErrorResponse(c, utils.OrderLogicErrCode, utils.OrderLogicErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, res{OrderId: orderId}, "success")
		}
	}
}

type respOrder struct {
	FromAddr      string `json:"from_addr"`
	RecipientAddr string `json:"recipient_addr"`
	Amount        string `json:"amount"`
	OrderType     int    `json:"order_type"`
	State         int    `json:"state"` // 订单状态, 0: pending; 1.完成；2.失败; 3. 超时
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("id")
		// 通过order id查询订单详情
		id, err := strconv.ParseUint(orderId, 10, 32)
		if err != nil {
			log.Errorf("strconv.ParseUint(orderId,10,32) error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		o, err := (&models.TacOrder{Model: gorm.Model{ID: uint(id)}}).GetOrder()
		if err != nil {
			log.Errorf("查询失败, err: %v", err)
			serializer.ErrorResponse(c, utils.OrderFindErrCode, utils.OrderFindErrMsg, err.Error())
			return
		}
		respOrder := respOrder{
			FromAddr:      o.FromAddr,
			RecipientAddr: o.RecipientAddr,
			Amount:        utils.UnitConversion(o.Amount, 8, 6),
			OrderType:     o.OrderType,
			State:         o.State,
		}
		serializer.SuccessResponse(c, respOrder, "success")
	}
}

type tacParams struct {
	OrderType  int    `json:"order_type" binding:"required"` // orderType == 1 表示拉取以太坊转tt的订单，为2则相反
	Address    string `json:"address" binding:"required"`
	StartIndex uint   `json:"start_index"`
	Limit      uint   `json:"limit" binding:"required"`
}
type tacResp struct {
	CreatedAt int64  `json:"created_at"`
	Amount    string `json:"amount"` // pala
	State     int    `json:"state"`  // 订单状态, 0: pending; 1.完成；2.失败; 3. 超时
}

// BatchGetTacOrder
func BatchGetTacOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params tacParams
		err := c.ShouldBind(&params)
		if err != nil {
			log.Errorf("get batch tac order error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		orders, err := new(models.TacOrder).GetBatchTacOrder(params.OrderType, params.Address, params.StartIndex, params.Limit)
		if err != nil {
			log.Errorf("flashChange order get batch error: %v", err)
			serializer.ErrorResponse(c, utils.TacOrderGetBatchErrCode, utils.TacOrderGetBatchErrMsg, err.Error())
			return
		} else {
			var resp []tacResp
			for _, o := range orders {
				r := tacResp{
					CreatedAt: o.CreatedAt.Unix(),
					Amount:    utils.UnitConversion(o.Amount, 8, 6),
					State:     o.State,
				}
				resp = append(resp, r)
			}
			serializer.SuccessResponse(c, resp, "success")
		}
	}
}
