package cmd

import (
	_ "embed"
	"github.com/leancodebox/GooseForum/app/service/bbsinit"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "bbs:init:db",
		Short: "bbs数据库初始化",
		Run:   runBbsinitDB,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}

func runBbsinitDB(_ *cobra.Command, _ []string) {
	bbsinit.DBInit()
}
