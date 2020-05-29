package models

import (
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
)

type TacOrder struct {
	gorm.Model
	FromAddr      string
	RecipientAddr string
	Amount        string
	OrderType     int
	State         int    // 订单状态, 0: pending; 1.完成；2.失败; 3. 超时
	SendTxHash    string // 跨链转账申请者发送给中间地址的交易hash
	ReceiveTxHash string // 中间地址发送给申请者的交易hash
	CollectionId  uint   // collection 表的外键
}

func (TacOrder) TableName() string {
	return "tac_order"
}

// Create
func (o *TacOrder) Create() error {
	if err := DB.Create(o).Error; err != nil {
		return err
	}
	return nil
}

// Exist 判断是否存在相同的订单
func (o *TacOrder) Exist(FromAddr, Amount string, OrderType int, State int) (*TacOrder, bool) {
	var ff TacOrder
	err := DB.Where("from_addr = ? AND amount = ? AND order_type = ? AND state = ?", FromAddr, Amount, OrderType, State).First(&ff).Error
	return &ff, !(err == gorm.ErrRecordNotFound)
}

// HasPendingOrder 存在pending中的同类型的订单
func (o *TacOrder) HasPendingOrder(FromAddr string, OrderType int, State int) bool {
	var ff TacOrder
	err := DB.Where("from_addr = ? AND order_type = ? AND state = ?", FromAddr, OrderType, State).First(&ff).Error
	return !(err == gorm.ErrRecordNotFound)
}

// GetOrder
func (o *TacOrder) GetOrder() (*TacOrder, error) {
	oo := TacOrder{}
	err := DB.Where(o).Last(&oo).Error
	return &oo, err
}

func (o *TacOrder) Update(oo TacOrder) error {
	return DB.Model(o).Updates(oo).Error
}

// 删除记录
func (o *TacOrder) Delete(orderId uint) error {
	return DB.Where("id = ?", orderId).Delete(&TacOrder{}).Error
}

// GetBatchTacOrder orderType == 1 表示拉取以太坊转tt的订单，为2则相反
func (o *TacOrder) GetBatchTacOrder(orderType int, fromAddress string, page uint, limit uint) ([]*TacOrder, int, error) {
	// 获取总的记录
	total := 0
	err := DB.Model(TacOrder{}).Where("from_addr = ? and order_type = ?", fromAddress, orderType).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit

	var orders []*TacOrder
	err = DB.Where("from_addr = ? and order_type = ?", fromAddress, orderType).Order("id desc").Limit(limit).Offset(offset).Find(&orders).Error
	if err != nil {
		log.Errorf("get batch by operate address err: %v", err)
		return nil, 0, err
	}
	return orders, total, nil
}

// GetTacOrdersByState 获取所有的state状态的order
func (o *TacOrder) GetTacOrdersByState(state int) ([]*TacOrder, error) {
	var orders []*TacOrder
	err := DB.Where("state = ?", state).Find(&orders).Error
	if err != nil {
		log.Errorf("get batch tac order by state err: %v", err)
		return nil, err
	}
	return orders, nil
}
