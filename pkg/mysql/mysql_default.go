package mysql

import (
	"cccc/pkg/config"
)

var mysqlCfg *MysqlConfig
var mysqlStore *StoreClient

func init() {
	mysqlCfg = &MysqlConfig{}
	mysqlStore = &StoreClient{}
	mysqlCfg.addRedepInit(mysqlStore.init)
	config.Default.Regist("mysql", mysqlCfg) // 注册config.Default默认对象
}

func Handle() *StoreClient {
	return mysqlStore
}
