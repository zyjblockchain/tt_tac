package models

import "testing"

func TestFlashChangeOrder_Exist(t *testing.T) {
	dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	InitDB(dsn)

	b := new(FlashChangeOrder).Exist("0x7AC954Ed6c2d96d48BBad405aa1579C828409f59", "0xD1Df5b185198F3c6Da74e93B36b7E29523c265F0", "0x03332638A6b4F5442E85d6e6aDF929Cd678914f1", 2)
	t.Log(b)
}
