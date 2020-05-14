package models

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"strings"
	"testing"
)

func TestGetOrder(t *testing.T) {
	dsn := "tac_user:tac_good123@tcp(rm-j6c49n23e4d07l8ijmo.mysql.rds.aliyuncs.com:3306)/tac_db?charset=utf8&parseTime=True&loc=Local"
	InitDB(dsn)
	from := strings.ToLower("0x7AC954Ed6c2d96d48BBad405aa1579C828409f59")
	ord, err := (&Order{FromAddr: from}).GetOrder()
	t.Log(err)
	t.Log(ord.ID)
	t.Log(*ord)
}
