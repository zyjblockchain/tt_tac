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
	State         int  // 订单状态, 0: pending; 1.完成；2.失败; 3. 超时
	CollectionId  uint // collection 表的外键
}

// Create
func (o *TacOrder) Create() error {
	if err := DB.Create(o).Error; err != nil {
		return err
	}
	return nil
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

// GetBatchTacOrder orderType == 1 表示拉取以太坊转tt的订单，为2则相反
func (o *TacOrder) GetBatchTacOrder(orderType int, fromAddress string, startId uint, limit uint) ([]*TacOrder, error) {
	var orders []*TacOrder
	err := DB.Where("from_addr = ? and order_type = ?", fromAddress, orderType).Order("id desc").Limit(limit).Offset(startId).Find(&orders).Error
	if err != nil {
		log.Errorf("get batch by operate address err: %v", err)
		return nil, err
	}
	return orders, nil
}
