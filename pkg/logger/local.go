package logger

import (
	"cccc/pkg/config"
	"context"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LocalLogger struct {
	log    *log.Logger
	level  KindLevel
	prefix string
	module string
	cfg    *LoggerConfig
	pool   sync.Pool
	buf    chan string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewLocalLogger(ss *config.SystemSignal, cfg *LoggerConfig) LogfI {
	ll := &LocalLogger{}
	ll.init(ss, cfg)
	return ll
}
func (l *LocalLogger) asyncLog() {
	for {
		select {
		case n := <-l.buf:
			l.log.Print(n)
		case <-l.ctx.Done():
			for i, s_i := 0, len(l.buf); i < s_i; i++ {
				n := <-l.buf
				l.log.Print(n)
			}
			return
		}
	}
}
func (ll *LocalLogger) init(ss *config.SystemSignal, cfg *LoggerConfig) {
	if ll.cancel != nil {
		ll.cancel()
	}
	fileName := cfg.LocalPath + cfg.Name + "." + cfg.Suffix
	wLoger := &lumberjack.Logger{
		Filename:   fileName,
		MaxBackups: cfg.Count,
		MaxSize:    cfg.Size,
		LocalTime:  true,
	}
	ll.cfg = cfg
	ll.pool.New = NewLogUnit
	ll.log = log.New(wLoger, cfg.AppName+" ", log.LstdFlags)
	ll.level = cfg.Level
	ll.buf = make(chan string, 300)
	ll.ctx, ll.cancel = context.WithCancel(ss.ContextExit())
	go ll.asyncLog()
}
func (l *LocalLogger) dumpLog(n *LogUnit) {
	minCallerDepth := 3
	maxCallerDepth := minCallerDepth + int(l.cfg.Depth) + 1
	pcs := make([]uintptr, maxCallerDepth)
	depth := runtime.Callers(minCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		n.FuncStack = append(n.FuncStack, CallStack{
			Line:     frame.Line,
			Function: frame.Function,
			File:     frame.File,
		})
		if !more {
			break
		}
	}
	switch l.cfg.Format {
	case OUTPUT_FMT_JSON:
		l.buf <- n.EncodeJson(l.cfg.Depth)
	case OUTPUT_FMT_TEXT:
		l.buf <- n.EncodeText(l.cfg.Depth)
	default:
		l.buf <- n.EncodeText(l.cfg.Depth)
	}
}
func (l *LocalLogger) getLogUnit() *LogUnit {
	n := l.pool.Get().(*LogUnit)
	n.AppName = l.cfg.AppName
	n.Prefix = l.prefix
	n.Timestamp = time.Now().UnixNano()
	n.Module = l.module
	return n
}
func (l *LocalLogger) putLogUnit(n *LogUnit) {
	n.Args = nil
	n.FuncStack = nil
	l.pool.Put(n)
}
func (l *LocalLogger) Debugf(format string, args ...interface{}) {
	if l.level > LOG_LEVEL_DEBUG {
		return
	}
	n := l.getLogUnit()
	n.Format = format
	n.Args = args
	n.Level = LOG_LEVEL_DEBUG.String()
	l.dumpLog(n)
	l.putLogUnit(n)
}
func (l *LocalLogger) Infof(format string, args ...interface{}) {
	if l.level > LOG_LEVEL_INFO {
		return
	}
	n := l.getLogUnit()
	n.Format = format
	n.Args = args
	n.Level = LOG_LEVEL_INFO.String()
	l.dumpLog(n)
	l.putLogUnit(n)
}
func (l *LocalLogger) Warnf(format string, args ...interface{}) {
	if l.level > LOG_LEVEL_WARN {
		return
	}
	n := l.getLogUnit()
	n.Format = format
	n.Args = args
	n.Level = LOG_LEVEL_WARN.String()
	l.dumpLog(n)
	l.putLogUnit(n)
}
func (l *LocalLogger) Errorf(format string, args ...interface{}) {
	if l.level > LOG_LEVEL_ERROR {
		return
	}
	n := l.getLogUnit()
	n.Format = format
	n.Args = args
	n.Level = LOG_LEVEL_ERROR.String()
	l.dumpLog(n)
	l.putLogUnit(n)
}
func (l *LocalLogger) Fatalf(format string, args ...interface{}) {
	n := l.getLogUnit()
	n.Format = format
	n.Args = args
	n.Level = LOG_LEVEL_FATAL.String()
	l.dumpLog(n)
	l.putLogUnit(n)
}
func (l *LocalLogger) Level(level string) {
	l.level = toLevel(level)
}

func (l *LocalLogger) Sublog(module string, args ...string) LogfI {
	ls := &LocalLogger{
		log:    l.log,
		prefix: strings.Join(args, "."),
		level:  l.level,
		module: module,
	}
	return ls
}
