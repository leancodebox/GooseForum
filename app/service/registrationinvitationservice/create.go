package registrationinvitationservice

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/invitations"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail               = errors.New("invalid invitation email")
	ErrMailUnavailable            = errors.New("invitation email is not configured")
	ErrInvitationNotFound         = errors.New("invitation not found")
	ErrInvitationEmailUnavailable = errors.New("invitation email cannot be retried")
)

// Create creates an invitation and, when requested, enqueues its email in the same DB transaction.
func Create(createdBy uint64, maxUses int, expiresAt *time.Time, recipient string) (invitations.Entity, string, bool, error) {
	recipient = strings.TrimSpace(recipient)
	if recipient != "" {
		address, err := mail.ParseAddress(recipient)
		if err != nil || address.Address != recipient {
			return invitations.Entity{}, "", false, ErrInvalidEmail
		}
		if err := mailservice.CheckConfigured(); err != nil {
			return invitations.Entity{}, "", false, ErrMailUnavailable
		}
	}

	var entity invitations.Entity
	var raw string
	err := dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		var err error
		entity, raw, err = invitations.CreateWithDB(tx, createdBy, maxUses, expiresAt)
		if err != nil || recipient == "" {
			return err
		}
		taskID, err := mailservice.AddToQueueWithDB(tx, mailservice.EmailTask{
			To: recipient, Token: raw, Type: mailservice.EmailTypeRegistrationInvitation,
		})
		if err != nil {
			return err
		}
		entity.EmailTaskId = &taskID
		return tx.Model(&invitations.Entity{}).Where("id = ?", entity.Id).Update("email_task_id", taskID).Error
	})
	return entity, raw, recipient != "" && err == nil, err
}

// RetryEmail requeues only a failed registration invitation task linked to the invitation.
func RetryEmail(invitationID uint64) error {
	if invitationID == 0 {
		return ErrInvitationNotFound
	}
	var invitation invitations.Entity
	if err := dbconnect.Connect().First(&invitation, invitationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}
	if invitation.EmailTaskId == nil {
		return ErrInvitationEmailUnavailable
	}
	retried, err := taskQueue.Retry(*invitation.EmailTaskId, mailservice.EmailTypeRegistrationInvitation)
	if err != nil {
		return err
	}
	if !retried {
		return ErrInvitationEmailUnavailable
	}
	return nil
}
