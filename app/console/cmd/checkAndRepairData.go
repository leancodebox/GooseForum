package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCollection"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/storage/model/articleLike"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "checkAndRepairData",
		Short: "检查和修复数据",
		Run:   runCheckAndRepairData,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	// cmd.Flags().StringP("param", "p", "value", "--param=x | -p x")
	appendCommand(cmd)
}

func runCheckAndRepairData(cmd *cobra.Command, args []string) {
	// param, _ := cmd.Flags().GetString("param")
	fmt.Println("检查用户")
	var userStartId uint64 = 0
	limit := 333
	for {
		userList := users.QueryById(userStartId, limit)
		for _, userItem := range userList {
			if userStartId < userItem.Id {
				userStartId = userItem.Id
			}
			userSt := userStatistics.Get(userItem.Id)
			if userSt.UserId == 0 {
				fmt.Println("用户统计信息不存在，开始初始化")
				userSt.UserId = userItem.Id
				userSt.LastActiveTime = &userItem.UpdatedAt
			}
			userSt.ArticleCount = cast.ToUint(articles.GetUserCount(userItem.Id))
			fmt.Println("获取文章总量", userSt.ArticleCount)
			userSt.ReplyCount = cast.ToUint(reply.GetUserCount(userItem.Id))
			fmt.Println("获取评论总量", userSt.ReplyCount)
			userSt.LikeReceivedCount = cast.ToUint(articleLike.GetLikeReceivedCount(userItem.Id))
			fmt.Println("获取接收到的点赞", userSt.LikeReceivedCount)
			userSt.LikeGivenCount = cast.ToUint(articleLike.GetLikeGivenCount(userItem.Id))
			fmt.Println("获送出点赞数量", userSt.LikeGivenCount)
			userSt.FollowingCount = cast.ToUint(userFollow.GetFollowingCount(userItem.Id)) //关注数
			fmt.Println("关注数", userSt.LikeGivenCount)
			userSt.FollowerCount = cast.ToUint(userFollow.GetFollowerCount(userItem.Id)) //粉丝数
			fmt.Println("粉丝数", userSt.LikeGivenCount)
			userSt.CollectionCount = cast.ToUint(articleCollection.GetCollectionCount(userItem.Id))
			fmt.Println("获取收藏数量", userSt.CollectionCount)
			userStatistics.SaveOrCreateById(&userSt)
			fmt.Println(userSt.UserId, "保存成功")
		}
		if len(userList) < limit {
			break
		}
	}
	fmt.Println("数据计算完毕")

}
