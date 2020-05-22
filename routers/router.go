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
		v1.POST("/get_eth_pala_receive", controllers.GetPalaReceiveRecord())
	}
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
