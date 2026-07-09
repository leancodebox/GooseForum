package pointservice

import "testing"

func TestPointsActionCode(t *testing.T) {
	tests := []struct {
		value PointsAction
		code  string
	}{
		{value: PointsActionInit, code: "init"},
		{value: PointsActionTopicPublished, code: "topic_published"},
		{value: PointsActionPostCreated, code: "post_created"},
		{value: PointsAction(99), code: "unknown"},
	}

	for _, tt := range tests {
		if got := tt.value.Code(); got != tt.code {
			t.Fatalf("Code() = %q, want %q", got, tt.code)
		}
	}
}
