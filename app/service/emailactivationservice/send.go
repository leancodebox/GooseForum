package emailactivationservice

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/tokenservice"
)

func SendActivationEmail(userEntity *users.EntityComplete) error {
	token, err := tokenservice.GenerateActivationTokenByUser(*userEntity)
	if err != nil {
		slog.Debug("生成激活邮件 Token 失败", "userId", userEntity.Id, "email", userEntity.Email, "err", err)
		return err
	}

	err = mailservice.AddToQueue(mailservice.EmailTask{
		To:       userEntity.Email,
		Username: userEntity.Username,
		Token:    token,
		Type:     "activation",
		Locale:   userEntity.Locale,
	})
	if err != nil {
		slog.Debug("激活邮件任务入队失败", "userId", userEntity.Id, "email", userEntity.Email, "err", err)
		return err
	}
	slog.Debug("激活邮件任务入队成功", "userId", userEntity.Id, "email", userEntity.Email)
	return nil
}
