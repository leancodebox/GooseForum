package api

import (
	"reflect"
	"testing"
)

func TestAdminTopicPayloadsDoNotExposeLegacyTopicType(t *testing.T) {
	if _, ok := reflect.TypeOf(TopicInfoAdminVo{}).FieldByName("Type"); ok {
		t.Fatal("TopicInfoAdminVo should not expose legacy topic type")
	}
	if _, ok := reflect.TypeOf(TopicSourceVo{}).FieldByName("Type"); ok {
		t.Fatal("TopicSourceVo should not expose legacy topic type")
	}
}
