package cmd

import (
	_ "embed"
	"github.com/leancodebox/GooseForum/app/service/databaseinit"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "init:db",
		Short: "GooseForum数据库初始化",
		Run:   runInitDB,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}

func runInitDB(_ *cobra.Command, _ []string) {
	databaseinit.DBInit()
}
