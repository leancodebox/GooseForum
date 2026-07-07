package cmd

import (
	"fmt"
	"math/rand"

	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "mock-posts",
		Short: "Mock topic posts",
		Run:   runMockTopicPosts,
	}
	cmd.Flags().Uint64("topic-id", 0, "Only mock posts for the specified topic")
	cmd.Flags().Int("count", 0, "Post count for --topic-id mode")
	cmd.Flags().Int("topics", 1000, "How many latest topics to scan when --topic-id is empty")
	cmd.Flags().Int("max-per-topic", 5, "Random upper bound per topic when --topic-id is empty")
	appendCommand(cmd)
}

func runMockTopicPosts(cmd *cobra.Command, args []string) {

	fmt.Println("Starting to mock posts...")
	topicID, _ := cmd.Flags().GetUint64("topic-id")
	count, _ := cmd.Flags().GetInt("count")
	topicLimit, _ := cmd.Flags().GetInt("topics")
	maxPerTopic, _ := cmd.Flags().GetInt("max-per-topic")

	if topicLimit <= 0 {
		topicLimit = 1000
	}
	if maxPerTopic <= 0 {
		maxPerTopic = 5
	}

	// 1. Get the latest published topics.
	topicList, err := topics.GetLatestPublished(topicLimit)
	if err != nil {
		fmt.Printf("Failed to load topics: %v\n", err)
		return
	}
	if topicID > 0 {
		topic := topics.Get(topicID)
		if topic.Id == 0 {
			fmt.Printf("Topic %d not found.\n", topicID)
			return
		}
		topicList = []*topics.Entity{&topic}
		if count <= 0 {
			count = 60
		}
	}
	fmt.Printf("Found %d topics\n", len(topicList))

	if len(topicList) == 0 {
		fmt.Println("No topics found.")
		return
	}

	// 2. Get all users
	userList := users.All()
	userIds := lo.Map(userList, func(u *users.EntityComplete, _ int) uint64 {
		return u.Id
	})

	if len(userIds) == 0 {
		fmt.Println("No users found.")
		return
	}
	fmt.Printf("Found %d users\n", len(userIds))

	// 3. Loop topics and generate posts.
	totalPosts := 0
	for i, topic := range topicList {
		postCount := count
		if topicID == 0 {
			postCount = rand.Intn(maxPerTopic)
		}
		if postCount == 0 {
			continue
		}

		for j := range postCount {
			// Pick random user
			userId := lo.Sample(userIds)

			req := component.BetterRequest[controllers.CreatePostReq]{
				Params: controllers.CreatePostReq{
					TopicId: topic.Id,
					Content: fmt.Sprintf("Mock post content %d-%d for topic %d. This is generated for local topic timeline testing.", i, j, topic.Id),
				},
				UserId: userId,
			}
			resp := controllers.CreatePost(req)

			if resp.Data.Code != component.SUCCESS {
				fmt.Printf("Error creating post for topic %d: %s\n", topic.Id, resp.Data.MessageCode)
			} else {
				totalPosts++
			}
		}
	}

	fmt.Printf("Success: Created %d mock posts for %d topics.\n", totalPosts, len(topicList))
}
