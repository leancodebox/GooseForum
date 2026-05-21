package cmd

import (
	"fmt"
	"strconv"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "user:send-activation <userId>",
		Short: "按用户 ID 发送账号激活邮件",
		Args:  cobra.ExactArgs(1),
		RunE:  runSendActivationEmail,
	}
	appendCommand(cmd)
}

func runSendActivationEmail(_ *cobra.Command, args []string) error {
	userID, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %s", args[0])
	}

	userEntity, err := users.Get(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}
	if userEntity.Id == 0 {
		return fmt.Errorf("用户ID %d 不存在", userID)
	}
	if userEntity.Email == "" {
		return fmt.Errorf("用户 %s(ID: %d) 未设置邮箱", userEntity.Username, userEntity.Id)
	}

	token, err := tokenservice.GenerateActivationTokenByUser(userEntity)
	if err != nil {
		return fmt.Errorf("生成激活 Token 失败: %w", err)
	}
	if err = mailservice.SendActivationEmail(userEntity.Email, userEntity.Username, token); err != nil {
		return fmt.Errorf("发送激活邮件失败: %w", err)
	}

	fmt.Printf("激活邮件已发送: userId=%d username=%s email=%s\n", userEntity.Id, userEntity.Username, userEntity.Email)
	return nil
}
