package models

import "github.com/jinzhu/gorm"

type TxTransfer struct {
	gorm.Model
	SenderAddress   string
	ReceiverAddress string
	TokenAddress    string // token合约地址
	Amount          string
	TxHash          string
	GasPrice        string // 此交易的gas price记录
	TxStatus        int    // 交易状态：0. pending; 1. 成功；2. 失败; 3. 超时
	OwnChain        int    // 所属链
	ErrMsg          string `gorm:"type:text(0)"` // 如果交易失败，则记录失败信息
}

func (TxTransfer) TableName() string {
	return "tx_transfer"
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
