package models

import "github.com/jinzhu/gorm"

type FlashChangeOrder struct {
	gorm.Model
	OperateAddress   string
	FromTokenAddress string
	ToTokenAddress   string
	FromTokenAmount  string
	ToTokenAmount    string
	State            int // 1. pending，2. success 3. failed 4. timeout
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
