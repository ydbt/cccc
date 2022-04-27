package logger_test

import (
	"cccc/pkg/logger"
	"context"
	"testing"
	"time"
)

var configLocal logger.LoggerConfig

func init() {
	configLocal.Level = logger.LOG_LEVEL_DEBUG
	configLocal.Size = 500
	configLocal.Count = 100
	configLocal.Name = "local_demo"
	configLocal.Suffix = "log"
	configLocal.LocalPath = "."
}
func TestLocalDebug(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	log := logger.NewLocalLogger(ctx, &configLocal)
	log.Debugf("%s", "666")
	cancel()
	time.Sleep(time.Millisecond * 100)
}
