package forum

import (
	"reflect"
	"testing"
)

func TestPublishPagePayloadDoesNotExposeLegacyTopicType(t *testing.T) {
	if _, ok := reflect.TypeOf(PublishPageProps{}).FieldByName("Types"); ok {
		t.Fatal("PublishPageProps should not expose legacy topic types")
	}
	if _, ok := reflect.TypeOf(PublishTopicPayload{}).FieldByName("Type"); ok {
		t.Fatal("PublishTopicPayload should not expose legacy topic type")
	}
}
