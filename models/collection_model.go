package models

import "github.com/jinzhu/gorm"

type CollectionTx struct {
	gorm.Model
	From         string
	To           string
	TokenAddress string
	Amount       string
	ChainNetUrl  string // 所属链
	TxId         uint   // tx表的外键
	IsValid      int    // 是否有效订单的收集, 0: false; 1. true
	ExtraInfo    string // 额外的信息
}

func (CollectionTx) TableName() string {
	return "collection_tx"
}

func (c *CollectionTx) Create() error {
	return DB.Create(c).Error
}

func (c *CollectionTx) Get() (*CollectionTx, error) {
	tt := CollectionTx{}
	err := DB.Where(c).Last(&tt).Error
	return &tt, err
}

func (c *CollectionTx) Update(cc CollectionTx) error {
	return DB.Model(c).Updates(cc).Error
}
