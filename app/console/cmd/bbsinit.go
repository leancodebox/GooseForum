package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/Users"

	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "bbs:init",
		Short: "bbs初始化",
		Run:   runBbsinit,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}

func runBbsinit(_ *cobra.Command, _ []string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	adminUsername := "t_admin"

	adminUser := Users.GetByUsername(adminUsername)
	if adminUser == nil {
		userEntity := Users.MakeUser(adminUsername, "123456", "admin@admin.com")
		err := Users.Create(userEntity)
		if err != nil {
			fmt.Println("账号创建失败，失败原因：", err)
		}
		fmt.Println(userEntity)
	}

}
