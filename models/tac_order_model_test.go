package models

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"testing"
)

func TestGetOrder(t *testing.T) {
	// dsn := "tac_user:tac_good123@tcp(rm-j6c49n23e4d07l8ijmo.mysql.rds.aliyuncs.com:3306)/tac_db?charset=utf8&parseTime=True&loc=Local"
	dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	InitDB(dsn)
	// from := strings.ToLower("0x7AC954Ed6c2d96d48BBad405aa1579C828409f59")
	// ord, err := (&TacOrder{FromAddr: from}).GetOrder()
	// t.Log(err)
	// t.Log(ord.ID)
	// t.Log(*ord)
	tacOrds, err := new(TacOrder).GetTacOrdersByState(0)
	t.Log(err)
	for _, o := range tacOrds {
		t.Log(o.ID)
		t.Log(o.CollectionId)
	}
}
