package mailservice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"testing"
)

func TestValidateConfigurationUsesEffectiveSender(t *testing.T) {
	config := pageConfig.MailSettingsConfig{EnableMail: true, SmtpHost: "smtp.example.com", SmtpPort: 587, SmtpUsername: "forum@example.com"}
	if err := validateConfiguration(config); err != nil {
		t.Fatalf("SMTP username sender fallback rejected: %v", err)
	}
	config.FromEmail = "not an email"
	if err := validateConfiguration(config); err == nil {
		t.Fatal("invalid configured sender accepted")
	}
}
