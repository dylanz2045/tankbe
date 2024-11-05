/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"user_mgt/common/pg_conn"
	usermgt "user_mgt/user_mgt"
	"user_mgt/utils"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var Config = utils.SetConfig()

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.Initlogger()
		if err != nil {
			fmt.Printf("logger init error:%v", err)
			os.Exit(-1)
		}
		if ok := pg_conn.Init(); !ok {
			utils.Logger.Fatalf("初始化数据库失败")
			os.Exit(-1)
		}

		//这里是启动两个模块之间的功能的
		usermgt.Run()

		// 启动http服务
		err = http.ListenAndServe(utils.AllConfig.ServerHost+":"+utils.AllConfig.ServerPort, nil)
		if err != nil {
			zap.L().Sugar().Errorf("http service failed to start: %s", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
