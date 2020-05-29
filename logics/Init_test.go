package logics

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zyjblockchain/tt_tac/models"
	"testing"
)

func TestInitTacOrderState(t *testing.T) {
	dsn := "tac_user:NwHJhkcTKHmDr2RZ@tcp(223.27.39.183:3306)/tac_db?charset=utf8mb4&parseTime=True&loc=Local"
	// dsn := "wallet_user_dev:NwHJhkcTKHmDr2RZ@tcp(rm-j6c49n23e4d07l8ijmo.mysql.rds.aliyuncs.com:3306)/kross_wallet_dev?charset=utf8mb4&parseTime=True&loc=Local"
	models.InitDB(dsn)
	InitTacOrderState()
}
