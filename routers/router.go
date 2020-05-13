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
		// 创建跨链转账的订单
		v1.POST("/apply_order", controllers.ApplyOrder())

	}
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
