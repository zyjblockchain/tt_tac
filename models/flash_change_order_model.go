package models

import (
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
)

type FlashChangeOrder struct {
	gorm.Model
	OperateAddress   string
	FromTokenAddress string // usdt token address
	ToTokenAddress   string // pala token address
	FromTokenAmount  string // usdt token amount
	ToTokenAmount    string // pala token amount
	TradePrice       string // 闪兑的兑换价格
	State            int    // 0. pending，1. success 2. failed 3. timeout
	SendTxId         uint   // 闪兑usdt发送的交易表id
	ReceiveTxId      uint   // 闪兑pala接收的交易表id
	CollectionId     uint   // collection 表的外键
}

func (FlashChangeOrder) TableName() string {
	return "flash_change_order"
}

func (f *FlashChangeOrder) Create() error {
	return DB.Create(f).Error
}

// 删除记录
func (f *FlashChangeOrder) Delete(orderId uint) error {
	return DB.Where("id = ?", orderId).Delete(&FlashChangeOrder{}).Error
}

func (f *FlashChangeOrder) Get() (*FlashChangeOrder, error) {
	tt := FlashChangeOrder{}
	err := DB.Where(f).Last(&tt).Error
	return &tt, err
}

func (f *FlashChangeOrder) Exist(operateAddr, fromTokenAddr, toTokenAddr string, state int) bool {
	var ff FlashChangeOrder
	err := DB.Where("operate_address = ? AND from_token_address = ? AND	to_token_address = ? AND state = ?", operateAddr, fromTokenAddr, toTokenAddr, state).First(&ff).Error
	return !(err == gorm.ErrRecordNotFound)
}

func (f *FlashChangeOrder) Update(ff FlashChangeOrder) error {
	return DB.Model(f).Updates(ff).Error
}

// page 页数，从第一页开始，limit 每一页的数量
func (f *FlashChangeOrder) GetBatchFlashOrder(operateAddress string, page, limit uint) ([]*FlashChangeOrder, int, error) {
	// 获取总的记录
	total := 0
	err := DB.Model(FlashChangeOrder{}).Where("operate_address = ?", operateAddress).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	var orders []*FlashChangeOrder
	err = DB.Where("operate_address = ?", operateAddress).Order("id desc").Limit(limit).Offset(offset).Find(&orders).Error
	if err != nil {
		log.Errorf("get batch by operate address err: %v", err)
		return nil, 0, err
	}
	return orders, total, nil
}

// GetFlashOrdersByState 获取所有的state状态的order
func (f *FlashChangeOrder) GetFlashOrdersByState(state int) ([]*FlashChangeOrder, error) {
	var orders []*FlashChangeOrder
	err := DB.Where("state = ?", state).Find(&orders).Error
	if err != nil {
		log.Errorf("get batch flash order by state err: %v", err)
		return nil, err
	}
	return orders, nil
}
