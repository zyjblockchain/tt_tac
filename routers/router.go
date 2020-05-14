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
	}
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
