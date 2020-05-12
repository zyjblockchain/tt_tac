package serializer

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 响应
type Response struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
	Error  string      `json:"error"`
}

func SuccessResponse(c *gin.Context, result interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Status: 200,
		Data:   result,
		Msg:    msg,
	})
}

func ErrorResponse(c *gin.Context, errorCode int, msg, err string) {
	c.JSON(200, Response{
		Status: errorCode,
		Msg:    msg,
		Error:  err,
	})
}
