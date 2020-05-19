package models

import "github.com/jinzhu/gorm"

type FlashChangeOrder struct {
	gorm.Model
	OperateAddress   string
	FromTokenAddress string // usdt token address
	ToTokenAddress   string // pala token address
	FromTokenAmount  string // usdt token amount
	ToTokenAmount    string // pala token amount
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

func (f *FlashChangeOrder) Update(ff FlashChangeOrder) error {
	return DB.Model(f).Updates(ff).Error
}
