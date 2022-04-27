package redis

import (
	"cccc/pkg/config"
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisHandle struct {
	redis.Cmdable
	Subscribe func(ctx context.Context, channels ...string) *redis.PubSub
	Close     func() error
}

// RedisConfig redis客户端配置
type RedisConfig struct {
	Address   []string `cccc:"address"`
	User      string   `cccc:"user"`
	Password  string   `cccc:"password"`
	ModeType  string   `cccc:"modeType"` // redis服务模型：cluster, sentinel, master-slave, single-point
	Instances []func(*config.SystemSignal, *RedisConfig)
}

func (cfg *RedisConfig) addRedepInit(f func(*config.SystemSignal, *RedisConfig)) {
	cfg.Instances = append(cfg.Instances, f)
}
func (cfg *RedisConfig) RedepInit(ss *config.SystemSignal) {
	for _, f := range cfg.Instances {
		f(ss, cfg)
	}
}

// Normalize redis客户端配置格式化
func (cfg *RedisConfig) Normalize() {
	if cfg.ModeType == "" {
		cfg.ModeType = "cluster"
	}
}

type RedisClient struct {
	client *RedisHandle
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *RedisConfig
}

func (r *RedisClient) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

func NewRedisClient(ss *config.SystemSignal, cfg *RedisConfig) *RedisClient {
	r := &RedisClient{}
	r.init(ss, cfg)
	return r
}
func (r *RedisClient) init(ss *config.SystemSignal, cfg *RedisConfig) {
	if r.cancel != nil {
		r.cancel()
	}
	if r.client != nil {
		r.client.Close()
	}
	r.ctx, r.cancel = context.WithCancel(ss.ContextExit())
	r.cfg = cfg
	switch r.cfg.ModeType {
	case "sentinel":
		c := redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs: r.cfg.Address,
			Username:      r.cfg.User,
			Password:      r.cfg.Password,
		})
		r.client.Cmdable = c
		r.client.Subscribe = c.Subscribe
		r.client.Close = c.Close
	case "single-point":
		fallthrough
	case "master-slave":
		c := redis.NewClient(&redis.Options{
			Addr:     r.cfg.Address[0],
			Username: r.cfg.User,
			Password: r.cfg.Password,
		})
		r.client.Cmdable = c
		r.client.Subscribe = c.Subscribe
		r.client.Close = c.Close
	case "cluster":
		c := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:          r.cfg.Address,
			Username:       r.cfg.User,
			Password:       r.cfg.Password,
			RouteByLatency: false,
			RouteRandomly:  false,
		})
		r.client.Cmdable = c
		r.client.Subscribe = c.Subscribe
		r.client.Close = c.Close
	}
}
