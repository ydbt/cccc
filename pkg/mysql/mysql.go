package mysql

import (
	"cccc/pkg/config"
	"cccc/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlConfig struct {
	MaxConn   int    `cccc:"maxConn"`   // 连接池最大连接数
	MinConn   int    `cccc:"minConn"`   // 连接池最小连接数
	KeepAlive int    `cccc:"keepAlive"` // 连接不在使用空闲时长,单位秒
	User      string `cccc:"user"`      // 数据库用户
	Password  string `cccc:"password"`  // 数据库密码
	Address   string `cccc:"address"`   // 数据库地址 10.130.29.76:3306
	Database  string `cccc:"database"`  // 数据库schema
	Charset   string `cccc:"charset"`   // 客户端字符集
	Location  string `cccc:"location"`  // 时区
	Instances []func(*config.SystemSignal, *MysqlConfig)
}

func (cfg *MysqlConfig) addRedepInit(f func(*config.SystemSignal, *MysqlConfig)) {
	cfg.Instances = append(cfg.Instances, f)
}
func (cfg *MysqlConfig) RedepInit(ss *config.SystemSignal) {
	for _, f := range cfg.Instances {
		f(ss, cfg)
	}
}

func (cfg *MysqlConfig) Normalize() {
	if cfg.MaxConn <= 0 {
		cfg.MaxConn = 10
	} else if cfg.MaxConn > 1000 {
		cfg.MaxConn = 100
	}
	if cfg.MinConn <= 0 {
		cfg.MinConn = (cfg.MaxConn / 4) + 1
	} else if cfg.MinConn > cfg.MaxConn {
		cfg.MinConn = (cfg.MaxConn / 2) + 1
	}
	if cfg.KeepAlive <= 0 {
		cfg.KeepAlive = 12 * 60 * 60
	}
	if cfg.Location == "" {
		cfg.Location = "Asia/Shanghai"
	}
}

type StoreClient struct {
	handle *sql.DB
	cfg    *MysqlConfig
	ss     *config.SystemSignal
	ctx    context.Context
	cancel context.CancelFunc
}

func (h *StoreClient) Ping() error {
	return h.Ping()
}
func (h *StoreClient) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if ctx == nil {
		ctx = h.ctx
	}
	return h.handle.QueryContext(ctx, query, args...)
}

type ExecResultItem struct {
	Index  int
	Result sql.Result
	Err    error
}
type ExecResultSet struct {
	Err   error
	Items []ExecResultItem
}
type StoreExecuteI interface {
	Fields() []interface{}
}

func (h *StoreClient) Exec(ctx context.Context, query string, datas []StoreExecuteI) (ers ExecResultSet) {
	if ctx == nil {
		ctx = h.ctx
	}
	var stmt *sql.Stmt
	stmt, ers.Err = h.handle.PrepareContext(ctx, query)
	if ers.Err != nil {
		return ers
	}
	defer stmt.Close()
	for i, data := range datas {
		result, err := stmt.ExecContext(ctx, data.Fields()...)
		if err != nil {
			ers.Items = append(ers.Items, ExecResultItem{
				Err:    err,
				Result: result,
				Index:  i,
			})
		}
	}
	return
}

func NewMysqlClient(ss *config.SystemSignal, cfg *MysqlConfig) *StoreClient {
	c := &StoreClient{}
	c.init(ss, cfg)
	if c.Ping() != nil {
		return nil
	}
	return c
}

func (c *StoreClient) init(ss *config.SystemSignal, cfg *MysqlConfig) {
	if c.cancel != nil {
		c.cancel()
	}
	c.cfg = cfg
	c.ss = ss
	c.ctx, c.cancel = context.WithCancel(ss.ContextExit())
	var err error
	c.handle, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&loc=%s&parseTime=True",
		cfg.User,
		cfg.Password,
		cfg.Address,
		cfg.Database,
		cfg.Charset,
		url.QueryEscape(cfg.Location)))
	if err != nil {
		logger.Fatalf("%v", err)
		c.handle = nil
	}
	c.handle.SetMaxIdleConns(cfg.MaxConn)
	c.handle.SetMaxIdleConns(cfg.MinConn)
	c.handle.SetConnMaxLifetime(time.Duration(cfg.KeepAlive) * time.Second)
}
