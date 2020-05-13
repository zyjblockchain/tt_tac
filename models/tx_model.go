package models

import "github.com/jinzhu/gorm"

type TxTransfer struct {
	gorm.Model
	SenderAddress   string
	ReceiverAddress string
	TokenAddress    string // token合约地址
	Amount          string
	TxHash          string
	TxStatus        int    // 交易状态：0. pending; 1. 成功；2. 失败; 3. 超时
	OwnChain        int    // 所属链
	ErrMsg          string // 如果交易失败，则记录失败信息
}

func (t *TxTransfer) Create() error {
	return DB.Create(t).Error
}

func (t *TxTransfer) Get() (*TxTransfer, error) {
	tt := TxTransfer{}
	err := DB.Where(t).Last(&tt).Error
	return &tt, err
}

func (t *TxTransfer) Update(tt TxTransfer) error {
	return DB.Model(t).Updates(tt).Error
}
