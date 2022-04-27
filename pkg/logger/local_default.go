package logger

import (
	"cccc/pkg/config"
)

var defaultLog *LocalLogger
var defaultCfg *LoggerConfig

var (
	Debugf func(string, ...interface{})
	Infof  func(string, ...interface{})
	Warnf  func(string, ...interface{})
	Errorf func(string, ...interface{})
	Fatalf func(string, ...interface{})
	Sublog func(string, ...string) LogfI
	Level  func(string)
)

func init() {
	defaultLog = &LocalLogger{}
	defaultCfg = &LoggerConfig{}
	defaultCfg.addRedepInit(defaultLog.init)
	config.Default.Regist("log", defaultCfg) // 注册config.Default默认对象
	defaultLog.module = "GLOBAL"
	Debugf = defaultLog.Debugf
	Infof = defaultLog.Infof
	Warnf = defaultLog.Warnf
	Errorf = defaultLog.Errorf
	Fatalf = defaultLog.Fatalf
	Sublog = defaultLog.Sublog
	Level = defaultLog.Level
}

func SetDefault(l LogfI) {
	sl := l.Sublog("GLOBAL")
	Debugf = sl.Debugf
	Infof = sl.Infof
	Warnf = sl.Warnf
	Errorf = sl.Errorf
	Fatalf = sl.Fatalf
	Sublog = sl.Sublog
	Level = sl.Level
}
