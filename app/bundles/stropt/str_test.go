package stropt

import "testing"

func TestCaseHelpers(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		snake      string
		camel      string
		lowerCamel string
	}{
		{
			name:       "camel input",
			input:      "TopicComment",
			snake:      "topic_comment",
			camel:      "TopicComment",
			lowerCamel: "topicComment",
		},
		{
			name:       "snake input",
			input:      "topic_comment",
			snake:      "topic_comment",
			camel:      "TopicComment",
			lowerCamel: "topicComment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Snake(tt.input); got != tt.snake {
				t.Fatalf("Snake(%q) = %q, want %q", tt.input, got, tt.snake)
			}
			if got := Camel(tt.input); got != tt.camel {
				t.Fatalf("Camel(%q) = %q, want %q", tt.input, got, tt.camel)
			}
			if got := LowerCamel(tt.input); got != tt.lowerCamel {
				t.Fatalf("LowerCamel(%q) = %q, want %q", tt.input, got, tt.lowerCamel)
			}
		})
	}
}

func TestPluralHelpers(t *testing.T) {
	if got := Plural("user"); got != "users" {
		t.Fatalf("Plural(user) = %q, want users", got)
	}
	if got := Singular("users"); got != "user" {
		t.Fatalf("Singular(users) = %q, want user", got)
	}
}
