package cmd

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "fixUserActivity",
		Short: "修复用户行为记录 (userActivities)",
		Run:   runFixUserActivity,
	}
	appendCommand(cmd)
}

func runFixUserActivity(cmd *cobra.Command, args []string) {
	fmt.Println("开始修复用户行为记录...")

	batchSize := 100

	// 1. 修复注册记录
	fmt.Println("正在同步注册记录...")
	offset := 0
	for {
		list, _ := users.GetAll(offset, batchSize)
		if len(list) == 0 {
			break
		}
		for _, u := range list {
			_ = userActivities.RecordWithTime(u.Id, userActivities.ActionSignUp, userActivities.SubjectUser, u.Id, "Joined the forum", u.CreatedAt)
		}
		offset += len(list)
		fmt.Printf("已处理 %d 条注册记录\n", offset)
	}

	// 2. 修复发帖记录
	fmt.Println("正在同步发帖记录...")
	offset = 0
	for {
		list, _ := articles.GetAllSimple(offset, batchSize)
		if len(list) == 0 {
			break
		}
		for _, a := range list {
			_ = userActivities.RecordWithTime(a.UserId, userActivities.ActionPost, userActivities.SubjectTopic, a.Id, a.Title, a.CreatedAt)
		}
		offset += len(list)
		fmt.Printf("已处理 %d 条发帖记录\n", offset)
	}

	// 3. 修复评论/回复记录
	fmt.Println("正在同步评论记录...")
	offset = 0
	for {
		list, _ := reply.GetAll(offset, batchSize)
		if len(list) == 0 {
			break
		}
		for _, r := range list {
			_ = userActivities.RecordWithTime(r.UserId, userActivities.ActionComment, userActivities.SubjectTopic, r.ArticleId, r.Content, r.CreatedAt)
		}
		offset += len(list)
		fmt.Printf("已处理 %d 条评论记录\n", offset)
	}

	// 4. 修复点赞记录
	fmt.Println("正在同步点赞记录...")
	offset = 0
	articleMap := make(map[uint64]string)
	for {
		list, _ := articleLike.GetAll(offset, batchSize)
		if len(list) == 0 {
			break
		}
		for _, l := range list {
			if l.Status == 0 {
				continue
			}
			title, ok := articleMap[l.ArticleId]
			if !ok {
				article := articles.GetSimple(l.ArticleId)
				title = article.Title
				articleMap[l.ArticleId] = title
			}
			_ = userActivities.RecordWithTime(l.UserId, userActivities.ActionLike, userActivities.SubjectTopic, l.ArticleId, title, l.CreatedAt)
		}
		offset += len(list)
		fmt.Printf("已处理 %d 条点赞记录\n", offset)
	}

	// 5. 修复关注记录
	fmt.Println("正在同步关注记录...")
	offset = 0
	for {
		list, _ := userFollow.GetAll(offset, batchSize)
		if len(list) == 0 {
			break
		}
		for _, f := range list {
			if f.Status == 0 {
				continue
			}
			_ = userActivities.RecordWithTime(f.UserId, userActivities.ActionFollow, userActivities.SubjectUser, f.FollowUserId, "Followed a user", f.CreatedAt)
		}
		offset += len(list)
		fmt.Printf("已处理 %d 条关注记录\n", offset)
	}

	fmt.Println("用户行为记录修复完成！")
}
