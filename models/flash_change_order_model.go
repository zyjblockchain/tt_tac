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

func (f *FlashChangeOrder) Create() error {
	return DB.Create(f).Error
}

func (f *FlashChangeOrder) Get() (*FlashChangeOrder, error) {
	tt := FlashChangeOrder{}
	err := DB.Where(f).Last(&tt).Error
	return &tt, err
}

func (f *FlashChangeOrder) Exist(operateAddr, fromTokenAddr, toTokenAddr string, state int) bool {
	var ff FlashChangeOrder
	err := DB.Where("operate_address = ? AND from_token_address = ? AND	to_token_address = ? AND state = ?", operateAddr, fromTokenAddr, toTokenAddr, state).First(&ff).Error
	log.Errorf("////: %v", err)
	return !(err == gorm.ErrRecordNotFound)
}

func (f *FlashChangeOrder) Update(ff FlashChangeOrder) error {
	return DB.Model(f).Updates(ff).Error
}

func (f *FlashChangeOrder) GetBatchFlashOrder(operateAddress string, startId, limit uint) ([]*FlashChangeOrder, error) {
	var orders []*FlashChangeOrder
	err := DB.Where("operate_address = ?", operateAddress).Order("id desc").Limit(limit).Offset(startId).Find(&orders).Error
	if err != nil {
		log.Errorf("get batch by operate address err: %v", err)
		return nil, err
	}
	return orders, nil
}
