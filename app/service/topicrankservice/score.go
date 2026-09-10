package topicrankservice

import (
	"time"
)

// Signals contains only counters and timestamps already persisted on topics.
type Signals struct {
	Likes, Replies uint64
	LastPostedAt   *time.Time
}

// Scores use thousandths of a point (30 points = 30000). The lookup table
// approximates ln(1+n) at 1e6 precision; scoring itself uses only integers.
const scorePrecision int64 = 1000000

var log1p = [...]int64{
	0, 693147, 1098612, 1386294, 1609438, 1791759, 1945910, 2079442, 2197225, 2302585,
	2397895, 2484907, 2564949, 2639057, 2708050, 2772589, 2833213, 2890372, 2944439, 2995732,
	3044522, 3091042, 3135494, 3178054, 3218876, 3258097, 3295837, 3332205, 3367296, 3401197,
	3433987, 3465736, 3496508, 3526361, 3555348, 3583519, 3610918, 3637586, 3663562, 3688879,
	3713572, 3737670, 3761200, 3784190, 3806662, 3828641, 3850148, 3871201, 3891820, 3912023,
	3931826, 3951244, 3970292, 3988984, 4007333, 4025352, 4043051, 4060443, 4077537, 4094345,
	4110874, 4127134, 4143135, 4158883, 4174387, 4189655, 4204693, 4219508, 4234107, 4248495,
	4262680, 4276666, 4290459, 4304065, 4317488, 4330733, 4343805, 4356709, 4369448, 4382027,
	4394449, 4406719, 4418841, 4430817, 4442651, 4454347, 4465908, 4477337, 4488636, 4499810,
	4510860, 4521789, 4532599, 4543295, 4553877, 4564348, 4574711, 4584967, 4595120, 4605170,
	4615121,
}

func compressed(n, cap int64) int64 {
	return log1p[min(max(n, 0), cap)] * scorePrecision / log1p[cap]
}

type timeStep struct {
	until time.Duration
	score int64
}

var freshnessSteps = []timeStep{
	{6 * time.Hour, 30}, {12 * time.Hour, 24}, {24 * time.Hour, 18},
	{48 * time.Hour, 12}, {72 * time.Hour, 8}, {7 * 24 * time.Hour, 4},
}

var activitySteps = []timeStep{
	{time.Hour, 20}, {6 * time.Hour, 12}, {12 * time.Hour, 6}, {24 * time.Hour, 2},
}

// Score performs fixed-cost integer arithmetic over one topic's saved values.
func Score(s Signals, published, now time.Time) (int64, *time.Time) {
	score := 6*compressed(int64(min(s.Likes, 50)), 50) + 14*compressed(int64(min(s.Replies, 100)), 100)
	fresh, next := timeScore(published, now, freshnessSteps)
	score += fresh * scorePrecision
	// last_posted_at also exists on topics without replies; require reply_count.
	if s.Replies > 0 && s.LastPostedAt != nil {
		activity, expiry := timeScore(*s.LastPostedAt, now, activitySteps)
		score += activity * scorePrecision
		if expiry != nil && (next == nil || expiry.Before(*next)) {
			next = expiry
		}
	}
	return (score + 500) / 1000, next
}

func timeScore(at, now time.Time, steps []timeStep) (int64, *time.Time) {
	for _, step := range steps {
		boundary := at.Add(step.until)
		if now.Before(boundary) {
			return step.score, &boundary
		}
	}
	return 0, nil
}
