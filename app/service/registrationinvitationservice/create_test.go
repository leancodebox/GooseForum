package registrationinvitationservice

import (
	"errors"
	"testing"
)

func TestCreateRejectsInvalidRecipientBeforeDatabaseWork(t *testing.T) {
	_, _, queued, err := Create(1, 1, nil, "not-an-email")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("Create() error = %v, want ErrInvalidEmail", err)
	}
	if queued {
		t.Fatal("Create() reported an email as queued after validation failure")
	}
}
