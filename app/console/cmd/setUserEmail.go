package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
	"github.com/spf13/cobra"
	"log/slog"
)

func init() {
	cmd := &cobra.Command{
		Use:   "user:email",
		Short: "发送邮件测试",
		Run:   runUserEmail,
		//Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	cmd.Flags().Uint64("userId", 1, "input userId --userId=1")
	appendCommand(cmd)

}

func runUserEmail(cmd *cobra.Command, _ []string) {
	userId, err := cmd.Flags().GetUint64("userId")
	if err != nil {
		fmt.Println(err)
		return
	}
	userEntity, _ := users.Get(userId)

	token, err := tokenservice.GenerateActivationToken(userEntity.Id, userEntity.Email)
	if err != nil {
		fmt.Println(err)
		return
	}
	//// 发送激活邮件
	//err = mailservice.SendActivationEmail(userEntity.Email, userEntity.Username, token)
	//if err != nil {
	//	fmt.Println(err)
	//	slog.Error("发送激活邮件失败", "error", err)
	//}
	err = mailservice.SendV2(userEntity.Email, userEntity.Username, token)
	if err != nil {
		fmt.Println(err)
		slog.Error("发送激活邮件失败", "error", err)
	}

}
