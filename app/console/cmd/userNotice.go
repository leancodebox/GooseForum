package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "userNotice",
		Short: "",
		Run:   runUserNotice,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	// cmd.Flags().StringP("param", "p", "value", "--param=x | -p x")
	appendCommand(cmd)
}

func runUserNotice(cmd *cobra.Command, args []string) {
	// param, _ := cmd.Flags().GetString("param")
	fmt.Println("success")

	// 根据用户循环遍历
	// 判断 userStatistics 的最后活跃时间
	// 根据 eventNotification , 判断用户是否存在7天内未读的通知
	// 判断 kvstore 记录的上次发送时间。
	// 综合判断后合并需要提醒的内容，发送邮件通知。
	// 这里不用真发送。打印需要发送的内容即可。

}
