package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
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
