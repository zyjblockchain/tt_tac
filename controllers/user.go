package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
	"strings"
)

type addr struct {
	Address string `json:"address"`
}

// 创建用户
func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.CreateUser
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("CreateUser should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		address, err := logic.CreateUser()
		if err != nil {
			log.Errorf("create user logic err: %v", err)
			serializer.ErrorResponse(c, utils.UserCreateErrCode, utils.UserCreateErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, addr{Address: address}, "success")
		}
	}
}

// 导入用户
func LeadUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.LeadUser
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("LeadUser should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		address, err := logic.LeadUser()
		if err != nil {
			log.Errorf("create user logic err: %v", err)
			serializer.ErrorResponse(c, utils.UserLeadErrCode, utils.UserLeadErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, addr{Address: address}, "success")
		}
	}
}

type rep struct {
	Private string `json:"private"`
}

// 导出私钥
func ExportPrivate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.Export
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("ExportPrivate should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		// logic
		private, err := logic.ExportPrivate()
		if err != nil {
			log.Errorf("ExportPrivate logic err: %v", err)
			serializer.ErrorResponse(c, utils.ExportPrivateErrCode, utils.ExportPrivateErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, rep{Private: private}, "success")
		}
	}
}

// 修改支付密码
func ModifyPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.ModifyPassword
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("ExportPrivate should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		err = logic.ModifyPwd()
		if err != nil {
			log.Errorf("ModifyPwd logic err: %v", err)
			serializer.ErrorResponse(c, utils.ModifyPasswordErrCode, utils.ModifyPasswordErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, nil, "success")
		}
	}
}

// CheckPassword 密码校验接口
func CheckPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.CheckPassword
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("CheckPassword should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		err, check := logic.CheckPwd()
		if err != nil {
			serializer.ErrorResponse(c, utils.CheckPasswordErrCode, utils.CheckPasswordErrMsg, "密码校验失败")
			return
		} else {
			serializer.SuccessResponse(c, check, "success")
		}
	}
}

// 拉取用户eth_Token 的收款记录
func GetEthTokenTxRecords(tokenSymbol string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.TokenTxsReceiveRecord
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("GetEthTokenTxRecords should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}
		// logic
		tokenAddress := ""
		decimal := 0
		if strings.ToUpper(tokenSymbol) == "PALA" {
			tokenAddress = "0xd20fb5cf926dc29c88f64725e6f911f40f7bf531" // 以太坊正式网上的pala合约地址
			decimal = 8
		} else if strings.ToUpper(tokenSymbol) == "USDT" {
			tokenAddress = "0xdac17f958d2ee523a2206206994597c13d831ec7" // 以太坊正式网上的usdt合约地址
			decimal = 6
		}

		records, err := logic.GetEthTokenTxRecord(tokenAddress, decimal)
		if err != nil {
			log.Errorf("GetEthTokenTxRecord logic err: %v", err)
			serializer.ErrorResponse(c, utils.GetEthTokenTxRecordErrCode, utils.GetEthTokenTxRecordErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, records, "success")
		}
	}
}

// GetEthReceiveRecords 拉取地址的eth收款记录
func GetEthReceiveRecords() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.EthTxsRecord
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("GetEthReceiveRecords should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		// logic
		records, err := logic.GetEthTxsRecord()
		if err != nil {
			log.Errorf("GetEthTxsRecord logic err: %v", err)
			serializer.ErrorResponse(c, utils.GetEthTxRecordErrCode, utils.GetEthTxRecordErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, records, "success")
		}
	}
}
