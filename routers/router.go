package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/zyjblockchain/tt_tac/middleware"
	"os"
)

func NewRouter(addr string) {
	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.Default()
	// 中间件
	r.Use(middleware.Cors())

	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
