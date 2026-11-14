package eventhandlers

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

func TestShouldPublishVisiblePost(t *testing.T) {
	tests := []struct {
		name         string
		topic        topics.Entity
		post         posts.Entity
		wasPublished bool
		want         bool
	}{
		{name: "first visible reply on published topic", topic: topics.Entity{Status: 1}, post: posts.Entity{ProcessStatus: 0}, want: true},
		{name: "unpublished topic", topic: topics.Entity{Status: 0}, post: posts.Entity{ProcessStatus: 0}, want: false},
		{name: "already published reply", topic: topics.Entity{Status: 1}, post: posts.Entity{ProcessStatus: 0}, wasPublished: true, want: false},
		{name: "still hidden reply", topic: topics.Entity{Status: 1}, post: posts.Entity{ProcessStatus: 1}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ShouldPublishVisiblePost(test.topic, test.post, test.wasPublished); got != test.want {
				t.Fatalf("ShouldPublishVisiblePost() = %v, want %v", got, test.want)
			}
		})
	}
}
