package models

import "github.com/jinzhu/gorm"

type SendTransfer struct {
	gorm.Model
	From         string
	To           string
	Amount       string
	TokenAddress string
	TxHash       string
	OwnChain     int    // 所属链
	CoinType     int    // 币种，1表示eth和tt主网币, 2表示pala币，3表示USDT币
	TxStatus     int    // 交易状态：0. pending; 1. 成功；2. 失败; 3. 超时
	ErrMsg       string `gorm:"type:text(0)"` // 如果交易失败，则记录失败信息
}

func (SendTransfer) TableName() string {
	return "send_transfer"
}

func (t *SendTransfer) Create() error {
	return DB.Create(t).Error
}

func (t *SendTransfer) Get() (*SendTransfer, error) {
	tt := SendTransfer{}
	err := DB.Where(t).Last(&tt).Error
	return &tt, err
}

func (t *SendTransfer) Update(tt SendTransfer) error {
	return DB.Model(t).Updates(tt).Error
}
