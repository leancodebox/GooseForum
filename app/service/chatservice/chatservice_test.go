package chatservice_test

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/chat/imUserChatConfigs"
	"github.com/leancodebox/GooseForum/app/models/chat/messages"
	"github.com/leancodebox/GooseForum/app/service/chatservice"
)

func TestMarkReadRequiresConversationMembership(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&imUserChatConfigs.Entity{}, &messages.Entity{}); err != nil {
		t.Fatal(err)
	}
	for i, tc := range []struct {
		name    string
		caller  uint64
		deleted bool
		target  string
		allowed bool
	}{
		{name: "nonparticipant", caller: 33},
		{name: "deleted membership", caller: 22, deleted: true},
		{name: "anonymous", caller: 0},
		{name: "missing conversation", caller: 22, target: "missing"},
		{name: "zero conversation", caller: 22, target: "zero"},
		{name: "participant", caller: 22, allowed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			convID := uint64(970000 + i)
			senderID, readerID := uint64(11+i*100), uint64(22+i*100)
			callerID := tc.caller
			if callerID != 0 {
				callerID += uint64(i * 100)
			}
			t.Cleanup(func() {
				db.Where("conv_id = ?", convID).Delete(&messages.Entity{})
				db.Where("conv_id = ?", convID).Delete(&imUserChatConfigs.Entity{})
				imUserChatConfigs.InvalidateConversationAccess(callerID, convID)
				imUserChatConfigs.InvalidateConversationAccess(callerID, convID+1000)
			})
			configs := []imUserChatConfigs.Entity{
				{UserId: senderID, PeerId: readerID, ConvId: convID, UnreadCount: 1},
				{UserId: readerID, PeerId: senderID, ConvId: convID, UnreadCount: 1},
			}
			if tc.deleted {
				configs[1].IsDeleted = 1
			}
			if err := db.Create(&configs).Error; err != nil {
				t.Fatal(err)
			}
			msgs := []messages.Entity{
				{ConvId: convID, SenderId: senderID, Content: "incoming message"},
				{ConvId: convID, SenderId: readerID, Content: "outgoing message"},
			}
			if err := db.Create(&msgs).Error; err != nil {
				t.Fatal(err)
			}
			targetID := convID
			switch tc.target {
			case "missing":
				targetID += 1000
			case "zero":
				targetID = 0
			}
			if err := chatservice.MarkRead(callerID, targetID); (err == nil) != tc.allowed {
				t.Errorf("MarkRead error = %v, allowed = %v", err, tc.allowed)
			}
			for index, original := range msgs {
				var stored messages.Entity
				if err := db.First(&stored, original.Id).Error; err != nil {
					t.Fatal(err)
				}
				wantRead := 0
				if tc.allowed && index == 0 {
					wantRead = 1
				}
				if stored.IsRead != wantRead {
					t.Errorf("message %d isRead = %d, want %d", stored.Id, stored.IsRead, wantRead)
				}
			}
			for _, original := range configs {
				var stored imUserChatConfigs.Entity
				if err := db.First(&stored, original.Id).Error; err != nil {
					t.Fatal(err)
				}
				wantUnread := uint(1)
				if tc.allowed && stored.UserId == callerID {
					wantUnread = 0
				}
				if stored.UnreadCount != wantUnread {
					t.Errorf("user %d unreadCount = %d, want %d", stored.UserId, stored.UnreadCount, wantUnread)
				}
			}
		})
	}
}
