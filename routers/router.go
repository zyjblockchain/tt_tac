package routers

import (
	"github.com/gin-gonic/gin"
	"os"
)

func NewRouter(addr string) {
	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.Default()
	// 中间件

	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
