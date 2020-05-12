package main

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/routers"
)

func main() {
	// 1. 初始化日志级别、格式、是否保存到文件
	log.Setup(log.LevelDebug, true, true)
	// 2. 数据库链接
	dsn := "tac_user:tac_good123@tcp(rm-j6c49n23e4d07l8ijmo.mysql.rds.aliyuncs.com:3306)/tac_db?charset=utf8&parseTime=True&loc=Local"
	models.InitDB(dsn)
	// 3. 启动服务
	routers.NewRouter(":3000")
}
