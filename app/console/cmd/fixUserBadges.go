package cmd

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/models/forum/userActivities"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "fixUserBadges",
		Short: "根据用户行为记录回填历史徽章",
		Run:   runFixUserBadges,
	}
	appendCommand(cmd)
}

func runFixUserBadges(cmd *cobra.Command, args []string) {
	fmt.Println("开始回填用户徽章...")

	const limit = 200
	var startId uint64
	processed := 0
	granted := 0

	for {
		userList := users.QueryById(startId, limit)
		if len(userList) == 0 {
			break
		}
		for _, user := range userList {
			if user == nil {
				continue
			}
			startId = user.Id
			granted += backfillUserBadges(user.Id)
			processed++
		}
		fmt.Printf("已处理 %d 个用户，新增授予 %d 个徽章\n", processed, granted)
		if len(userList) < limit {
			break
		}
	}

	fmt.Printf("用户徽章回填完成，处理用户 %d 个，新增授予 %d 个徽章。\n", processed, granted)
}

func backfillUserBadges(userID uint64) int {
	actions := userActionCountMap(userID)
	codes := make([]string, 0, 9)

	if actions[int(userActivities.ActionPost)] >= 1 {
		codes = append(codes, badgeservice.CodeFirstPost)
	}
	if actions[int(userActivities.ActionPost)] >= 10 {
		codes = append(codes, badgeservice.CodeWriter10)
	}
	if actions[int(userActivities.ActionComment)] >= 1 {
		codes = append(codes, badgeservice.CodeFirstComment)
	}
	if actions[int(userActivities.ActionComment)] >= 50 {
		codes = append(codes, badgeservice.CodeCommenter50)
	}
	if actions[int(userActivities.ActionLike)] >= 1 {
		codes = append(codes, badgeservice.CodeFirstLikeGiven)
	}

	count := 0
	for _, code := range append(codes, receivedBadgeCandidates(userID)...) {
		granted, err := badgeservice.Grant(userID, code, "migration", "历史行为回填", 0)
		if err != nil {
			fmt.Printf("用户 %d 回填徽章 %s 失败: %v\n", userID, code, err)
			continue
		}
		if granted {
			count++
		}
	}
	return count
}

func userActionCountMap(userID uint64) map[int]int64 {
	result := make(map[int]int64)
	for _, item := range userActivities.CountActionsByUser(userID) {
		result[item.Action] = item.Count
	}
	return result
}

func receivedBadgeCandidates(userID uint64) []string {
	stats := userStatistics.Get(userID)
	codes := make([]string, 0, 4)
	if stats.LikeReceivedCount >= 10 {
		codes = append(codes, badgeservice.CodeLiked10)
	}
	if stats.LikeReceivedCount >= 100 {
		codes = append(codes, badgeservice.CodePopular100)
	}
	if stats.FollowerCount >= 1 {
		codes = append(codes, badgeservice.CodeFirstFollower)
	}
	if stats.FollowerCount >= 10 {
		codes = append(codes, badgeservice.CodeSocial10)
	}
	return codes
}
