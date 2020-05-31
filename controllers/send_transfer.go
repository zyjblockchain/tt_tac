package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
)

type tHash struct {
	TxHash string `json:"tx_hash"`
}

// SendTacTx 发送跨链转账
func SendTacTx() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.SendTacTx
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("ApplyOrder should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// 把传入的amount换算成最小单位
		logic.Amount = utils.FormatTokenAmount(logic.Amount, 8)
		log.Infof("发送跨链转账交易前端传入参数：Address：%s, amount: %s, orderType: %d, tacOrderId: %d", logic.Address, logic.Amount, logic.OrderType, logic.TacOrderId)

		// logic
		// 查看是否存在订单
		tacOrd, err := (&models.TacOrder{Model: gorm.Model{ID: logic.TacOrderId}}).GetOrder()
		if err != nil {
			// 表示不存在ord
			log.Errorf("跨链转账申请者发送交易到中间地址传入的orderId在tacOrder表中不存在。 db error: %v, tacOrderId: %d", err, logic.TacOrderId)
			serializer.ErrorResponse(c, utils.SendTacTxErrCode, utils.SendTacTxErrMsg, errors.New("跨链转账申请者发送交易到中间地址传入的orderId在tacOrder表中不存在").Error())
			return
		}
		txHash, err := logic.SendTacTx()
		if err != nil {
			// 发送跨链转账发送者转账到中转地址交易失败，则把跨链转账订单置位失败状态
			log.Errorf("跨链转账申请者发送跨链转账交易失败, 把申请跨链转账订单置位失败状态；error: %v", err)
			oo := &models.TacOrder{}
			oo.ID = logic.TacOrderId
			_ = oo.Update(models.TacOrder{State: 2}).Error()
			if err == utils.VerifyPasswordErr {
				// 返回密码校验失败的状态码给前端
				serializer.ErrorResponse(c, utils.CheckPasswordErrCode, utils.CheckPasswordErrMsg, "密码错误")
				return
			}
			serializer.ErrorResponse(c, utils.SendTacTxErrCode, utils.SendTacTxErrMsg, err.Error())
			return
		} else {
			// sendTxHash更新到tacOrder中
			_ = tacOrd.Update(models.TacOrder{SendTxHash: txHash})
			serializer.SuccessResponse(c, tHash{TxHash: txHash}, "success")
		}
	}
}

// SendPalaTransfer 发送pala交易
func SendPalaTransfer(chainTag int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.PalaTransfer
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("SendPalaTransfer should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// 把传入的amount换算成最小单位
		logic.Amount = utils.FormatTokenAmount(logic.Amount, 8)
		log.Infof("发送pala转账交易；chainTag: %d, from: %s, to: %s, amount: %s", chainTag, logic.FromAddress, logic.ToAddress, logic.Amount)

		// logic
		txHash, err := logic.SendPalaTx(chainTag)
		if err != nil {
			if err == utils.VerifyPasswordErr {
				serializer.ErrorResponse(c, utils.CheckPasswordErrCode, utils.CheckPasswordErrMsg, "密码错误")
				return
			}
			// 发送pala转账交易失败
			log.Errorf("发送pala转账交易失败；error: %v", err)
			serializer.ErrorResponse(c, utils.SendPalaTransferErrCode, utils.SendPalaTransferErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, tHash{TxHash: txHash}, "success")
		}
	}
}

// SendMainCoin 发送eth或者tt主网币交易
func SendMainCoin(chainTag int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.CoinTransfer
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("SendMainCoin should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		logic.Amount = utils.FormatTokenAmount(logic.Amount, 18)
		log.Infof("发送主网币转账交易；chainTag: %d, from: %s, to: %s, amount: %s", chainTag, logic.FromAddress, logic.ToAddress, logic.Amount)

		// logic
		txHash, err := logic.SendMainNetCoinTransfer(chainTag)
		if err != nil {
			if err == utils.VerifyPasswordErr {
				serializer.ErrorResponse(c, utils.CheckPasswordErrCode, utils.CheckPasswordErrMsg, "密码错误")
				return
			}
			log.Errorf("发送主网币转账交易；error: %v", err)
			serializer.ErrorResponse(c, utils.SendMainCoinTransferErrCode, utils.SendMainCoinTransferErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, tHash{TxHash: txHash}, "success")
		}
	}
}

// SendEthUsdtTransfer eth usdt转账
func SendEthUsdtTransfer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.EthUsdtTransfer
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("SendEthUsdtTransfer should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		// 把传入的amount换算成最小单位
		logic.Amount = utils.FormatTokenAmount(logic.Amount, 6)
		log.Infof("发送eth_usdt转账交易； from: %s, to: %s, amount: %s", logic.FromAddress, logic.ToAddress, logic.Amount)

		// logic
		txHash, err := logic.SendEthUsdtTransfer()
		if err != nil {
			if err == utils.VerifyPasswordErr {
				serializer.ErrorResponse(c, utils.CheckPasswordErrCode, utils.CheckPasswordErrMsg, "密码错误")
				return
			}

			// 发送usdt转账交易失败
			log.Errorf("发送usdt转账交易失败；error: %v", err)
			serializer.ErrorResponse(c, utils.SendUsdtTransferErrCode, utils.SendUsdtTransferErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, tHash{TxHash: txHash}, "success")
		}
	}
}

type SendTxRecord struct {
	Address string `json:"address" binding:"required"`
	Page    uint   `json:"page" binding:"required"`
	Limit   uint   `json:"limit"`
}

// 分页拉取地址下的pala发送交易记录
func GetSendTransferRecords(ownChain, coinType, decimal int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params SendTxRecord
		err := c.ShouldBind(&params)
		if err != nil {
			log.Errorf("get batch GetSendTransferRecords error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		if params.Limit == 0 {
			params.Limit = 5
		}
		records, total, err := new(models.SendTransfer).GetBatchSendTransfer(params.Address, ownChain, coinType, params.Page, params.Limit)
		if err != nil {
			log.Errorf("GetSendTransferRecords get batch error: %v", err)
			serializer.ErrorResponse(c, utils.GetSendTransferRecordsErrCode, utils.GetSendTransferRecordsErrMsg, err.Error())
			return
		} else {
			resp := make([]tacResp, 0, 0)
			for _, o := range records {
				r := tacResp{
					CreatedAt: o.CreatedAt.Unix(),
					Amount:    utils.UnitConversion(o.Amount, decimal, 6),
					State:     o.TxStatus,
				}
				resp = append(resp, r)
			}
			serializer.SuccessResponse(c, Result{Total: total, List: resp}, "success")
		}
	}
}
