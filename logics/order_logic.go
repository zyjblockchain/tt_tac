package logics

import (
	"errors"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
	"strings"
)

type Order struct {
	FromAddr      string `json:"fromAddr" binding:"required"`
	RecipientAddr string `json:"recipientAddr" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
	Password      string `json:"password" binding:"required"`
	OrderType     int    `json:"orderType" binding:"required"`
}

// CreateOrder 返回订单号
func (ord *Order) CreateOrder() (uint, error) {
	// 查看地址是否存在数据库
	u, err := new(models.User).GetUserByAddress(ord.FromAddr)
	if err != nil {
		log.Errorf("地址不存在. error: %v", err)
		return 0, err
	}
	// 验证password是否正确
	if !u.CheckPassword(ord.Password) {
		log.Errorf("密码有误")
		return 0, utils.VerifyPasswordErr
	}

	order := &models.TacOrder{
		FromAddr:      strings.ToLower(utils.FormatAddressHex(ord.FromAddr)),
		RecipientAddr: strings.ToLower(utils.FormatAddressHex(ord.RecipientAddr)),
		Amount:        ord.Amount,
		OrderType:     ord.OrderType,
		State:         0,
	}
	// 查看数据库中是否存在相同的订单
	exist := order.HasPendingOrder(order.FromAddr, order.OrderType, 0)
	if exist {
		// 数据库中存在
		log.Errorf("数据库中存在相同地址下的相同类型正在pending的跨链转账订单；FromAddr：%s, Amount: %s, OrderType: %d, State: %d", order.FromAddr, order.Amount, order.OrderType, order.State)
		return 0, errors.New("数据库中存在相同地址下的相同类型正在pending的跨链转账订单，请等待上一个跨链转账订单完成之后再重试")
	}
	// 保存到数据库
	err = order.Create()
	if err != nil {
		return 0, err
	} else {
		return order.ID, nil
	}
}
