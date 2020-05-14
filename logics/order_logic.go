package logics

import (
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
	"strings"
)

type Order struct {
	FromAddr      string `json:"fromAddr" binding:"required"`
	RecipientAddr string `json:"recipientAddr" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
	OrderType     int    `json:"orderType" binding:"required"`
}

// CreateOrder 返回订单号
func (ord *Order) CreateOrder() (uint, error) {
	order := &models.Order{
		FromAddr:      strings.ToLower(utils.FormatHex(ord.FromAddr)),
		RecipientAddr: strings.ToLower(utils.FormatHex(ord.RecipientAddr)),
		Amount:        ord.Amount,
		OrderType:     ord.OrderType,
		State:         0,
	}
	// 保存到数据库
	err := order.Create()
	if err != nil {
		return 0, err
	} else {
		return order.ID, nil
	}
}
