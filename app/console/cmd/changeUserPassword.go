package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "user:changePassword",
		Short: "更改用户密码",
		Run:   changeUserPassword,
		//Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	cmd.Flags().Uint64("userId", 1, "input userId --userId=1")
	cmd.Flags().String("password", "123456", "input password --password=123456")
	appendCommand(cmd)

}

func changeUserPassword(cmd *cobra.Command, _ []string) {
	userId, err := cmd.Flags().GetUint64("userId")
	password, err := cmd.Flags().GetString("password")
	if err != nil {
		fmt.Println(err)
		return
	}
	userEntity, _ := users.Get(userId)
	if userEntity.Id == 0 {
		fmt.Println("用户不存在")
		return
	}
	userEntity.SetPassword(password)
	r := users.Save(&userEntity)
	fmt.Println(r)
}
