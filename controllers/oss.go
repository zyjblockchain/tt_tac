package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/conf"
	"github.com/zyjblockchain/tt_tac/serializer"
	"github.com/zyjblockchain/tt_tac/utils"
	"math/big"
)

// EncryptoPrivate 对传入的私钥进行加密
type Priv struct {
	Private string `json:"private" binding:"required"`
}

func EncryptoPrivate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var p Priv
		err := c.ShouldBind(&p)
		if err != nil {
			log.Errorf("EncryptoPrivate should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		private := p.Private
		ePrivate, err := utils.EncryptPrivate(private)
		if err != nil {
			log.Errorf("EncryptPrivate  err: %v", err)
			serializer.ErrorResponse(c, utils.EncryptoPrivErrCode, utils.EncryptoPrivErrMsg, err.Error())
			return
		} else {
			serializer.SuccessResponse(c, ePrivate, "success")
		}
	}
}

// 获取跨链pala手续费
func GetTacPalaServiceCharge() gin.HandlerFunc {
	return func(c *gin.Context) {
		chr := charge{
			ToTtCharge:  utils.UnitConversion(conf.EthToTtPalaCharge.String(), 8, 8),
			ToEthCharge: utils.UnitConversion(conf.TtToEthPalaCharge.String(), 8, 8),
		}
		serializer.SuccessResponse(c, chr, "success")
	}
}

// 修改跨链转账收取一定量的pala作为手续费的接口
type charge struct {
	ToTtCharge  string `json:"to_tt_charge"`
	ToEthCharge string `json:"to_eth_charge"`
}

func ModifyTacPalaServiceCharge() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Infof("old charge; to_tt_charge: %s, to_eth_charge: %s", utils.UnitConversion(conf.EthToTtPalaCharge.String(), 8, 8), utils.UnitConversion(conf.TtToEthPalaCharge.String(), 8, 8))
		var newCharge charge
		err := c.ShouldBind(&newCharge)
		if err != nil {
			log.Errorf("ModifyTacPalaServiceCharge should binding error: %v", err)
			serializer.ErrorResponse(c, utils.VerifyParamsErrCode, utils.VerifyParamsErrMsg, err.Error())
			return
		}

		// 修改conf中的EthToTtPalaCharge和TtToEthPalaCharge
		if newCharge.ToTtCharge != "" {
			tc := utils.FormatTokenAmount(newCharge.ToTtCharge, 8) // 前端传的都是pala的最大单位，我们需要转换成pala的最小单位
			rr, b := new(big.Int).SetString(tc, 10)
			if !b {
				log.Errorf("new(big.Int).SetString(tc, 10)  err: %v", err)
				serializer.ErrorResponse(c, utils.ModifyTacPalaServiceChargeErrCode, utils.ModifyTacPalaServiceChargeErrMsg, err.Error())
				return
			}
			conf.EthToTtPalaCharge = rr
		}
		if newCharge.ToEthCharge != "" {
			tc := utils.FormatTokenAmount(newCharge.ToEthCharge, 8) // 前端传的都是pala的最大单位，我们需要转换成pala的最小单位
			rr, b := new(big.Int).SetString(tc, 10)
			if !b {
				log.Errorf("new(big.Int).SetString(tc, 10)  err: %v", err)
				serializer.ErrorResponse(c, utils.ModifyTacPalaServiceChargeErrCode, utils.ModifyTacPalaServiceChargeErrMsg, err.Error())
				return
			}
			conf.TtToEthPalaCharge = rr
		}
		log.Infof("new charge; to_tt_charge: %s, to_eth_charge: %s", utils.UnitConversion(conf.EthToTtPalaCharge.String(), 8, 8), utils.UnitConversion(conf.TtToEthPalaCharge.String(), 8, 8))
		serializer.SuccessResponse(c, nil, "success")
	}
}
