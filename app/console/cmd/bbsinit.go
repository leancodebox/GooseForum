package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/service/bbsinit"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "bbs:init",
		Short: "bbs初始化",
		Run:   runBBsinit,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}

func runBBsinit(_ *cobra.Command, _ []string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	bbsinit.DataInit()
}
