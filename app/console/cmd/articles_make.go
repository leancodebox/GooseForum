package cmd

import (
	"context"
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/Users"
	Articles2 "github.com/leancodebox/GooseForum/app/models/bbs/Articles"
	Comment2 "github.com/leancodebox/GooseForum/app/models/bbs/Comment"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "tool:articles_make",
		Short: "articles_make",
		Run:   runArticlesMake,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
	appendCommand(&cobra.Command{
		Use:   "tool:createAndDeleted",
		Short: "createAndDeleted",
		Run:   createAndDeleted,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})

	appendCommand(&cobra.Command{
		Use:   "tool:createAndUpdate",
		Short: "createAndUpdate",
		Run:   createAndUpdate,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}

func createAndUpdate(_ *cobra.Command, _ []string) {
	art := Articles2.Articles{UserId: 1, Content: `
你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好
`}
	Articles2.Save(&art)

	art.Content = "haohaohaohaohao"

	time.Sleep(time.Second * 3)

	Articles2.Save(&art)

	fmt.Println(art)
}

func createAndDeleted(_ *cobra.Command, _ []string) {
	art := Articles2.Articles{UserId: 1, Content: `
你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好
`}
	Articles2.Save(&art)

	Articles2.Delete(&art)

	fmt.Println(art)
}

func runArticlesMake(_ *cobra.Command, _ []string) {
	userEntity := Users.MakeUser(cast.ToString(time.Now().UnixMilli()), "123456", cast.ToString(time.Now())+"@qq.com")
	err := Users.Create(userEntity)
	if err != nil {
		fmt.Println("用户创建失败", err)
	}

	userList := Users.All()
	fmt.Print(userList)
	ctx := context.Background()

	ArticlesRep := Articles2.NewRep(&ctx)
	CommentRep := Comment2.NewRep(&ctx)
	for _, user := range userList {
		for i := 0; i < 10; i++ {

			art := Articles2.Articles{UserId: user.Id, Content: `
你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好
`}
			ArticlesRep.Save(&art)
			for _, cUser := range userList {
				comment := Comment2.Comment{UserId: cUser.Id, ArticleId: art.Id, Content: cUser.Username + "觉得不错"}
				CommentRep.Save(&comment)
				comment = Comment2.Comment{UserId: cUser.Id, ArticleId: art.Id, Content: cUser.Username + "觉得不错"}
				CommentRep.Save(&comment)
				comment = Comment2.Comment{UserId: cUser.Id, ArticleId: art.Id, Content: cUser.Username + "觉得不错"}
				CommentRep.Save(&comment)
			}
		}
	}
}
