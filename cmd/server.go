package cmd

import (
	"cccc/pkg/config"
	"cccc/pkg/logger"
	_ "cccc/pkg/toolset" // 初始化时区
	"cccc/src"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

var serverConfig, serverRunMode string
var serverCmd = &cobra.Command{
	Use:   `server`,
	Short: "微服务启动",
	Long:  "根据配置启动微服务",
	Run: func(cmd *cobra.Command, args []string) {
		config.Default.LoadConfigFile(serverConfig)
		go func() {
			chSG := make(chan os.Signal)
			signal.Notify(chSG, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
			sg := <-chSG
			config.Default.SS.WaitExit()
			serverRunMode = ""
			logger.Warnf("received signal:%d", sg)
			src.Kill()
		}()
		if serverRunMode != "daemon" && serverRunMode != "forever" {
			logger.Infof("foreground process beg")
			defer logger.Infof("foreground process end")
			src.Run(config.Default.SS)
			os.Exit(0)
		}
		if env := os.Getenv("CCCC_DAEMON_SWITCH"); env == "" || env == "0" {
			// 父进程启动子进程
			logger.Infof("Init SubProcess beg")
			defer logger.Infof("Init SubProcess end")
			appPath, _ := filepath.Abs(os.Args[0])
			cmd := exec.Command(appPath, os.Args[1:]...)
			cmd.Env = append(cmd.Env, "CCCC_DAEMON_SWITCH=1") // 设置环境变量用于区分子进程于父进程
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
			os.Exit(0)
		} else {
			// 子进程
			for serverRunMode == "forever" || !config.Default.SS.Exited() {
				logger.Infof("background process beg")
				src.Run(config.Default.SS)
				logger.Infof("background process end")
			}
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(&serverConfig, "config", "c", "config/cccc.yaml", "服务启动配置")
	serverCmd.Flags().StringVarP(&serverRunMode, "runmode", "r", "normal", "服务启动模式；normal:前台模式，daemon：后台模式，forever：守护模式")
	rootCmd.AddCommand(serverCmd)
}
