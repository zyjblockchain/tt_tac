package logics

import (
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
)

type Order struct {
	FromAddr      string `json:"fromAddr" binding:"required"`
	RecipientAddr string `json:"recipientAddr" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
	OrderType     int    `json:"orderType" binding:"required"`
}

func (ord *Order) CreateOrder() error {
	order := &models.Order{
		FromAddr:      utils.FormatHex(ord.FromAddr),
		RecipientAddr: utils.FormatHex(ord.RecipientAddr),
		Amount:        ord.Amount,
		OrderType:     ord.OrderType,
		State:         0,
	}
	// 保存到数据库
	err := order.Create()
	return err
}
