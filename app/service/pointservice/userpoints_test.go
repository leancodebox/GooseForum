package pointservice

import "testing"

func TestRewardPointsTypeString(t *testing.T) {
	tests := []struct {
		value RewardPointsType
		want  string
	}{
		{value: RewardPointsInit, want: "初始化"},
		{value: RewardPoints4WriteArticles, want: ""},
		{value: RewardPoints4Reply, want: ""},
		{value: RewardPointsType(99), want: ""},
	}

	for _, tt := range tests {
		if got := tt.value.String(); got != tt.want {
			t.Fatalf("String() = %q, want %q", got, tt.want)
		}
	}
}
