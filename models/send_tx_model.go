package models

import (
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
)

type SendTransfer struct {
	gorm.Model
	FromAddress  string
	ToAddress    string
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

// GetBatchSendTransfer page 页数，从第一页开始，limit 每一页的数量
func (t *SendTransfer) GetBatchSendTransfer(fromAddr string, ownChain, coinType int, page, limit uint) ([]*SendTransfer, int, error) {
	// 获取总的记录
	total := 0
	err := DB.Model(SendTransfer{}).Where("from_address = ? and own_chain = ? and coin_type = ?", fromAddr, ownChain, coinType).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	var orders []*SendTransfer
	err = DB.Where("from_address = ? and own_chain = ? and coin_type = ?", fromAddr, ownChain, coinType).Order("id desc").Limit(limit).Offset(offset).Find(&orders).Error
	if err != nil {
		log.Errorf("get batch SendTransfer err: %v", err)
		return nil, 0, err
	}
	return orders, total, nil
}
