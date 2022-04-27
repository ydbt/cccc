package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cccc",
	Short: "智能外呼微服务框架",
	Long:  "微服务统一启动入口",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cccc", args)
		cmd.Help()
	},
}

func Execute() {
	rootCmd.Execute()
}
