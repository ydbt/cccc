package redis

import (
	"cccc/pkg/config"
)

var redisCfg *RedisConfig
var redisClient *RedisClient

func init() {
	redisCfg = &RedisConfig{}
	redisClient = &RedisClient{}
	redisCfg.addRedepInit(redisClient.init)
	config.Default.Regist("redis", redisCfg) // 注册config.Default默认对象
}

func Handler() *RedisHandle {
	return redisClient.client
}
