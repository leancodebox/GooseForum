package postwriteservice

import (
	"errors"
	"testing"

	"github.com/leancodebox/GooseForum/app/service/trustlevelservice"
)

func TestReplyWritesRejectAttachments(t *testing.T) {
	ids := []uint64{401}
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			var err error
			if operation == "create" {
				_, err = Create(CreateInput{AttachmentIDs: &ids, UserID: 7, TopicID: 60, Content: "Reply with attachment"})
			} else {
				_, err = Update(UpdateInput{AttachmentIDs: &ids, UserID: 7, PostID: 61, Content: "Edited reply with attachment"})
			}
			var policy *trustlevelservice.PolicyError
			if !errors.As(err, &policy) || policy.Status != 400 || policy.Code != "common.request.invalidParams" {
				t.Fatalf("attachment reply error = %v, want invalid parameters", err)
			}
		})
	}
}
