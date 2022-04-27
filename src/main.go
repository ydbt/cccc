package src

import (
	"cccc/pkg/config"
	"cccc/pkg/logger"
	"fmt"
	"time"
)

func Run(ss *config.SystemSignal) {
	go func() {
		logger.Infof("[src.Run] beg")
		defer logger.Infof("[src.Run] end")
		for i := 0; i < 60; i++ {
			fmt.Println("[src.Run]", i)
			time.Sleep(time.Second)
		}
		ss.NotifyExit()
	}()
	ss.WaitExit()
}

func Kill() {
}

func Stop() {
}
