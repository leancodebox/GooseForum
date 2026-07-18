package topicunseenservice

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/service/kvstore"
)

const (
	trackingTTL       = 48 * time.Hour
	activeVisitWindow = 30 * time.Minute
	valueVersion      = byte(1)
	trackingValueSize = 17
	visitValueSize    = 9
	keyPrefix         = "topic-unseen:v1:"
)

type TopicActivity struct {
	TopicID      uint64
	LastPostedAt time.Time
}

type trackingState struct {
	LastActiveAt time.Time
	LastSeenAt   time.Time
}

// TouchUser advances the user's activity heartbeat without resolving topics.
func TouchUser(userID uint64, now time.Time) error {
	if userID == 0 {
		return nil
	}
	_, err := touchTracking(userID, now)
	return err
}

// Resolve returns per-topic unseen state and advances the user's activity heartbeat.
func Resolve(userID uint64, activities []TopicActivity, now time.Time) (map[uint64]bool, error) {
	result := make(map[uint64]bool, len(activities))
	if userID == 0 || len(activities) == 0 {
		return result, nil
	}
	tracking, err := touchTracking(userID, now)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(activities))
	for _, activity := range activities {
		if activity.TopicID > 0 {
			keys = append(keys, visitKey(userID, activity.TopicID))
		}
	}
	values, err := kvstore.GetManyBytes(keys)
	if err != nil {
		return nil, err
	}
	visited := make(map[uint64]time.Time, len(values))
	for _, activity := range activities {
		value, exists := values[visitKey(userID, activity.TopicID)]
		if !exists {
			continue
		}
		visitedAt, ok := decodeVisit(value)
		if ok {
			visited[activity.TopicID] = visitedAt
		}
	}
	return resolveUnseen(activities, visited, minLastSeenAt(tracking.LastSeenAt, now)), nil
}

// MarkVisited records a successful topic detail visit and refreshes user activity.
func MarkVisited(userID, topicID uint64, now time.Time) error {
	if userID == 0 || topicID == 0 {
		return nil
	}
	_, trackingErr := touchTracking(userID, now)
	visitErr := kvstore.UpdateBytes(visitKey(userID, topicID), trackingTTL, func(current []byte, exists bool) (kvstore.UpdateAction, []byte, error) {
		if exists {
			visitedAt, ok := decodeVisit(current)
			if ok && !now.After(visitedAt) {
				return kvstore.UpdateSet, current, nil
			}
		}
		return kvstore.UpdateSet, encodeVisit(now), nil
	})
	return errors.Join(trackingErr, visitErr)
}

func touchTracking(userID uint64, now time.Time) (trackingState, error) {
	var result trackingState
	err := kvstore.UpdateBytes(userKey(userID), trackingTTL, func(current []byte, exists bool) (kvstore.UpdateAction, []byte, error) {
		state := trackingState{}
		if exists {
			state, _ = decodeTracking(current)
		}
		result = nextTracking(state, now)
		return kvstore.UpdateSet, encodeTracking(result), nil
	})
	return result, err
}

func nextTracking(current trackingState, now time.Time) trackingState {
	if now.IsZero() {
		return current
	}
	if current.LastActiveAt.IsZero() {
		return trackingState{LastActiveAt: now, LastSeenAt: current.LastSeenAt}
	}
	if !now.After(current.LastActiveAt) {
		return current
	}
	next := current
	if now.Sub(current.LastActiveAt) > activeVisitWindow {
		next.LastSeenAt = current.LastActiveAt
	}
	next.LastActiveAt = now
	return next
}

func minLastSeenAt(lastSeenAt, now time.Time) time.Time {
	floor := now.Add(-trackingTTL)
	if lastSeenAt.After(floor) {
		return lastSeenAt
	}
	return floor
}

func resolveUnseen(activities []TopicActivity, visited map[uint64]time.Time, baseline time.Time) map[uint64]bool {
	result := make(map[uint64]bool, len(activities))
	for _, activity := range activities {
		if activity.TopicID == 0 || activity.LastPostedAt.IsZero() || !activity.LastPostedAt.After(baseline) {
			continue
		}
		visitedAt, exists := visited[activity.TopicID]
		result[activity.TopicID] = !exists || activity.LastPostedAt.After(visitedAt)
	}
	return result
}

func userKey(userID uint64) string {
	return keyPrefix + "user:" + strconv.FormatUint(userID, 10)
}

func visitKey(userID, topicID uint64) string {
	return fmt.Sprintf("%svisit:%d:%d", keyPrefix, userID, topicID)
}

func encodeTracking(state trackingState) []byte {
	value := make([]byte, trackingValueSize)
	value[0] = valueVersion
	binary.BigEndian.PutUint64(value[1:9], encodeTime(state.LastActiveAt))
	binary.BigEndian.PutUint64(value[9:17], encodeTime(state.LastSeenAt))
	return value
}

func decodeTracking(value []byte) (trackingState, bool) {
	if len(value) != trackingValueSize || value[0] != valueVersion {
		return trackingState{}, false
	}
	return trackingState{
		LastActiveAt: decodeTime(binary.BigEndian.Uint64(value[1:9])),
		LastSeenAt:   decodeTime(binary.BigEndian.Uint64(value[9:17])),
	}, true
}

func encodeVisit(visitedAt time.Time) []byte {
	value := make([]byte, visitValueSize)
	value[0] = valueVersion
	binary.BigEndian.PutUint64(value[1:9], encodeTime(visitedAt))
	return value
}

func decodeVisit(value []byte) (time.Time, bool) {
	if len(value) != visitValueSize || value[0] != valueVersion {
		return time.Time{}, false
	}
	return decodeTime(binary.BigEndian.Uint64(value[1:9])), true
}

func encodeTime(value time.Time) uint64 {
	if value.IsZero() {
		return 0
	}
	return uint64(value.UnixMilli())
}

func decodeTime(value uint64) time.Time {
	if value == 0 {
		return time.Time{}
	}
	return time.UnixMilli(int64(value)).UTC()
}
