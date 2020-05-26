package models

import (
	"testing"
)

func TestFlashChangeOrder_Exist(t *testing.T) {
	dsn := "wallet_user_dev:NwHJhkcTKHmDr2RZ@tcp(rm-j6c49n23e4d07l8ijmo.mysql.rds.aliyuncs.com:3306)/kross_wallet_dev?charset=utf8mb4&parseTime=True&loc=Local"
	InitDB(dsn)

	// b := new(FlashChangeOrder).Exist("0x7AC954Ed6c2d96d48BBad405aa1579C828409f59", "0xD1Df5b185198F3c6Da74e93B36b7E29523c265F0", "0x03332638A6b4F5442E85d6e6aDF929Cd678914f1", 2)
	// t.Log(b)
	var sendTxIds []uint
	DB.Model(FlashChangeOrder{}).Where("state = ?", 1).Pluck("send_tx_id", &sendTxIds)
	for _, v := range sendTxIds {
		t.Log(v)
		var tx = TxTransfer{}
		tx.ID = v
		DB.Select("gas_price").Take(&tx)
		t.Log(tx.GasPrice)
		t.Log(tx.Amount)
	}
}
