package component

import (
	"encoding/json"
	"testing"
)

func TestPageJSONContract(t *testing.T) {
	response := SuccessPage([]string{"first"}, 2, 10, 11)
	data, err := json.Marshal(response.Data.Result)
	if err != nil {
		t.Fatalf("marshal offset page: %v", err)
	}
	var offset map[string]any
	if err := json.Unmarshal(data, &offset); err != nil {
		t.Fatalf("decode offset page: %v", err)
	}
	if offset["page"] != float64(2) || offset["size"] != float64(10) || offset["total"] != float64(11) {
		t.Fatalf("offset metadata = %#v", offset)
	}
	if _, ok := offset["hasNext"]; ok {
		t.Fatal("offset page unexpectedly serialized hasNext=false")
	}

	cursorPage, err := json.Marshal(Page[string]{List: []string{"first"}, Page: 1, Size: 20, HasNext: true})
	if err != nil {
		t.Fatalf("marshal cursor page: %v", err)
	}
	var cursor map[string]any
	if err := json.Unmarshal(cursorPage, &cursor); err != nil {
		t.Fatalf("decode cursor page: %v", err)
	}
	if cursor["hasNext"] != true || cursor["page"] != float64(1) || cursor["size"] != float64(20) {
		t.Fatalf("cursor metadata = %#v", cursor)
	}
}
