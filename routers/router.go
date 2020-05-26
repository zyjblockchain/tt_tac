package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/tt_tac/controllers"
	"github.com/zyjblockchain/tt_tac/middleware"
	"os"
)

func NewRouter(addr string) {
	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.Default()
	// 跨域中间件
	r.Use(middleware.Cors())

	v1 := r.Group("/tac")
	{
		// 1. 创建跨链转账的订单, 返回订单id；
		v1.POST("/apply_order", controllers.ApplyOrder())
		// 2. 通过订单id查询订单详情 http://127.0.0.1:3000/tac/order/111
		v1.GET("/order/:id", controllers.GetOrder())
		// 3. 发送跨链转账交易
		v1.POST("/send_tac_tx", controllers.SendTacTx())
		// 4. 创建用户
		v1.POST("/create_wallet", controllers.CreateUser())
		// 5. 导入用户
		v1.POST("/lead_wallet", controllers.LeadUser())
		// 6. 导出私钥
		v1.POST("/export_private", controllers.ExportPrivate())
		// 7. 修改支付密码
		v1.POST("/modify_password", controllers.ModifyPassword())
		// 闪兑
		v2 := v1.Group("/exchange")
		{
			// 1. 以太坊上的usdt兑换eth_pala
			v2.POST("/eth_usdt_pala", controllers.FlashChange())
			// 2. 分页拉取地址下面闪兑订单列表
			v2.POST("/get_flash_orders", controllers.GetBatchOrderByAddress())
		}
		// 8. 获取地址的balance, tt链和eth上的balance
		v1.POST("/get_balance", controllers.GetBalance())
		// 9. 获取地址的token balance, tt链和eth上的token balance
		v1.POST("/get_token_balance", controllers.GetTokenBalance())
		// 10. 获取btcMax交易所上的erc20 pala价格,锚定的usdt
		v1.GET("/get_eth_pala_price", controllers.GetLatestPalaToUsdtPrice())
		// 11. 获取btcMax交易所上的eth价格,锚定的usdt
		v1.GET("/get_eth_price", controllers.GetLatestEthToUsdtPrice())
		// 12. 分页拉取跨链转账的订单记录
		v1.POST("/get_tac_orders", controllers.BatchGetTacOrder())
		// 13. 获取用户地址下的eth主网上pala接收记录
		v1.POST("/get_eth_pala_receive", controllers.GetEthTokenTxRecords("PALA"))
		// 14. 获取用户地址下的eth主网上usdt接收记录
		v1.POST("/get_eth_usdt_receive", controllers.GetEthTokenTxRecords("USDT"))
		// 15. 拉取地址下的eth的收款记录
		v1.POST("/get_eth_receive", controllers.GetEthReceiveRecords())
		// 16. 获取发送一笔以太坊token转账交易或者tt的token转账交易需要的gas fee
		v1.POST("/get_gas_fee", controllers.GetGasFee())

		// 内部管理接口
		// 1. 对私钥进行对称加密，用于配置中间地址的私钥加密
		v1.POST("/encrypto_private", controllers.EncryptoPrivate())
		// 2. 获取跨链转账扣除pala手续费
		v1.GET("/get_tac_charge", controllers.GetTacPalaServiceCharge())
		// 3. 修改跨链转账扣除pala手续费数量接口
		v1.POST("/modify_tac_charge", controllers.ModifyTacPalaServiceCharge())
		// 4. 获取闪兑中的pala价格的上浮比例
		v1.GET("/get_pala_price_change_rate", controllers.GetPalaPriceComeUpRate())
		// 5. 修改闪兑中的pala价格的上浮比例
		v1.POST("/modify_pala_price_change_rate", controllers.ModifyPalaPriceComeUpRate())
		// 6. 获取闪兑的交易gas消耗总量
		v1.GET("/get_flash_total_gas_fee", controllers.GetFlashTotalGasFee())
		// 7. 获取跨链转账的交易gas消耗总量
		v1.GET("/get_tac_total_gas_fee", controllers.GetTacTotalGasFee())
		// 8. 获取闪兑的usdt接收总量和pala发送总量
		v1.GET("/get_flash_pala_usdt_total", controllers.GetFlashUsdtAndPalaTotalAmount())
	}
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
