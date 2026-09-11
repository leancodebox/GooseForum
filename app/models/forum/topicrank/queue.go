// Package topicrank marks topics for asynchronous score updates.
package topicrank

import (
	"context"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
)

// Wakeups only nudges the local worker. The database remains the schedule;
// repeated notifications collapse into one buffered wakeup.
var wakeups = make(chan struct{}, 1)

func Wakeups() <-chan struct{} { return wakeups }

// TopicRankRequested carries values only, never a request context or transaction handle.
type TopicRankRequested struct {
	TopicID    uint64
	OccurredAt time.Time
}

// Notify uses the existing asynchronous event bus; it performs no database I/O.
func Notify(topicID uint64) {
	if topicID == 0 {
		return
	}
	eventbus.Publish(context.Background(), &TopicRankRequested{TopicID: topicID, OccurredAt: time.Now()})
}
