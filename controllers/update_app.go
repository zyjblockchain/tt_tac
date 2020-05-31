package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/logics"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
)

// GetAppVersion 获取当前app版本
func GetAppVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		serializer.SuccessResponse(c, logics.AppVersionInfo, "success")
	}
}

// CheckUpdate app 版本更新校验
func CheckUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		version, exist := c.GetQuery("version")
		if !exist {
			// 参数不存在
			log.Errorf("CheckUpdate 参数不存在")
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, "version参数不存在")
			return
		}
		resp := logics.CheckUpdate(version)
		serializer.SuccessResponse(c, &resp, "success")
	}
}

// SetAppVersion
func SetAppVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		var logic logics.SetVersionInfo
		err := c.ShouldBind(&logic)
		if err != nil {
			log.Errorf("SetAppVersion should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		// logic
		err = logic.SetAppVersionInfo()
		if err != nil {
			log.Errorf("设置app版本更新失败；error: %v", err)
			serializer.ErrorResponse(c, utils.SetAppVersionErrCode, utils.SetAppVersionErrMsg, err.Error())
			return
		}
		serializer.SuccessResponse(c, nil, "success")
	}
}
