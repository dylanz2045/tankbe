/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/spf13/cobra"
)

// htmltestCmd represents the htmltest command
var htmltestCmd = &cobra.Command{
	Use:   "htmltest",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", sayHello)
		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			fmt.Println("HTTP server failed,err:", err)
			return
		}
	},
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	// 解析指定文件生成模板对象
	tmpl, err := template.ParseFiles("./template.html")
	if err != nil {
		fmt.Println("create template failed, err:", err)
		return
	}
	// 定义数据
	data := struct {
		Title    string
		Name     string
		LoggedIn bool
	}{
		Title:    "Welcome Page",
		Name:     "Kimi",
		LoggedIn: false,
	}
	// 利用给定数据渲染模板，并将结果写入w
	tmpl.Execute(w, data)
}

func init() {
	rootCmd.AddCommand(htmltestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// htmltestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// htmltestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
