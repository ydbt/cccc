package logger

import (
	"cccc/pkg/config"
	"fmt"
	"strings"
)

type KindOutputFormat uint8

const (
	OUTPUT_FMT_TEXT KindOutputFormat = iota
	OUTPUT_FMT_JSON
)

type KindLevel uint

const (
	LOG_LEVEL_DEBUG KindLevel = iota // DEBUG
	LOG_LEVEL_INFO                   // INFO
	LOG_LEVEL_WARN                   // WARN
	LOG_LEVEL_ERROR                  // ERROR
	LOG_LEVEL_FATAL                  // FATAL
)

type LoggerConfig struct {
	Async      bool             `cccc:"aync"`       // 异步日志
	Format     KindOutputFormat `cccc:"formt"`      // 文件格式；0:文本，1:JSON
	Level      KindLevel        `cccc:"level"`      // 日志等级
	Depth      uint8            `cccc:"depth"`      // 函数调用栈深度
	Count      int              `cccc:"count"`      // 日志文件个数
	Size       int              `cccc:"size"`       // 单个日志文件大小，单位MB
	AppName    string           `cccc:"appName"`    // 服务名
	Name       string           `cccc:"name"`       // 日志文件名称
	Suffix     string           `cccc:"suffix"`     // 日志文件后缀
	LocalPath  string           `cccc:"localPath"`  // 本地日志路径
	RemotePath string           `cccc:"remotePath"` // 远程日志请求路径
	Instances  []func(*config.SystemSignal, *LoggerConfig)
}

func (cfg *LoggerConfig) addRedepInit(f func(*config.SystemSignal, *LoggerConfig)) {
	cfg.Instances = append(cfg.Instances, f)
}
func (cfg *LoggerConfig) RedepInit(ss *config.SystemSignal) {
	for _, f := range cfg.Instances {
		f(ss, cfg)
	}
}

// 配置初始化之后的校验
func (cfg *LoggerConfig) Normalize() {
	if cfg.Level < LOG_LEVEL_DEBUG || cfg.Level > LOG_LEVEL_FATAL {
		cfg.Level = LOG_LEVEL_DEBUG
	}
	if cfg.Count <= 0 {
		cfg.Count = 100
	} else if cfg.Count > 10000 {
		cfg.Count = 10000
	}
	if cfg.Size <= 0 {
		cfg.Size = 300
	} else if cfg.Size > 10240 {
		cfg.Size = 10240
	}
	if cfg.Name == "" {
		cfg.Name = "cccc"
	}
	if cfg.LocalPath == "" {
		cfg.LocalPath = "./"
	}
	if cfg.Suffix == "" {
		pos := strings.LastIndex(cfg.Name, ".")
		if pos == -1 {
			cfg.Suffix = "log"
		} else {
			cfg.Name = cfg.Name[:pos]
			cfg.Suffix = cfg.Name[pos+1:]
		}
	}
	if !strings.HasSuffix(cfg.LocalPath, "/") {
		cfg.LocalPath += "/"
	}
}

type LogfI interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	Sublog(string, ...string) LogfI
	Level(string)
}

func (k KindLevel) String() string {
	desc := "DEBUG"
	switch k {
	case LOG_LEVEL_DEBUG:
		desc = "DEBUG"
	case LOG_LEVEL_INFO:
		desc = "INFO"
	case LOG_LEVEL_WARN:
		desc = "WARN"
	case LOG_LEVEL_ERROR:
		desc = "ERROR"
	case LOG_LEVEL_FATAL:
		desc = "FATAL"
	}
	return desc
}
func toLevel(l string) KindLevel {
	ll := strings.ToUpper(l)
	k := LOG_LEVEL_DEBUG
	switch ll {
	case "DEBUG":
		k = LOG_LEVEL_DEBUG
	case "INFO":
		k = LOG_LEVEL_INFO
	case "WARN":
		k = LOG_LEVEL_WARN
	case "ERROR":
		k = LOG_LEVEL_ERROR
	case "FATAL":
		k = LOG_LEVEL_FATAL
	}
	return k
}

func NewLogger(cfg *LoggerConfig) *LogfI {
	return nil
}

type CallStack struct {
	Line     int
	Function string
	File     string
}
type LogUnit struct {
	Timestamp int64         `json:"timestamp"`   // 日志时间微秒
	Level     string        `json:"log_level"`   // 日志等级
	Module    string        `json:"module_name"` // 模块名
	AppName   string        `json:"app_name"`    // 服务名称
	Prefix    string        `json:"sub_prefix"`  // 前缀
	Format    string        `json:"-"`           // 格式化
	Args      []interface{} `json:"-"`
	FuncStack []CallStack   `json:"call_stack"` // 函数调用栈
}

func (lu *LogUnit) EncodeText(depth uint8) string {
	fs := ""
	for j, s_j, s_k := 0, int(depth+1), len(lu.FuncStack); j < s_j && j < s_k; j++ {
		fs += fmt.Sprintf("[%s] %s:%d", lu.FuncStack[j].Function, lu.FuncStack[j].File, lu.FuncStack[j].Line)
	}
	return fmt.Sprintln(lu.Level, lu.Prefix, fs, fmt.Sprintf(lu.Format, lu.Args...))
}
func (lu *LogUnit) EncodeJson(depth uint8) string {
	return ""
}

func NewLogUnit() interface{} {
	return &LogUnit{}
}
