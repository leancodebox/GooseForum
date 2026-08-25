package datamigration

import (
	"encoding/json"
	"testing"
)

func TestAddDefaultTopicLimit(t *testing.T) {
	updated, changed, err := addDefaultTopicLimit(`{"textControl":{"maxPostLength":50000}}`, 10)
	if err != nil || !changed {
		t.Fatalf("addDefaultTopicLimit() = %q, %v, %v", updated, changed, err)
	}
	var value map[string]map[string]int
	if err := json.Unmarshal([]byte(updated), &value); err != nil {
		t.Fatalf("decode updated config: %v", err)
	}
	if value["textControl"]["maxDailyTopicsPerUser"] != 10 {
		t.Fatalf("maxDailyTopicsPerUser = %d, want 10", value["textControl"]["maxDailyTopicsPerUser"])
	}
}

func TestAddDefaultTopicLimitPreservesExplicitValue(t *testing.T) {
	input := `{"textControl":{"maxDailyTopicsPerUser":0}}`
	updated, changed, err := addDefaultTopicLimit(input, 10)
	if err != nil || changed || updated != input {
		t.Fatalf("addDefaultTopicLimit() = %q, %v, %v", updated, changed, err)
	}
}
