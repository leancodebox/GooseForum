package mailservice

import (
	"strings"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestBuildEmailActionURL(t *testing.T) {
	got := buildEmailActionURL("https://example.com/", "/activate", "a/b+c=")
	want := "https://example.com/activate?token=a%2Fb%2Bc%3D"
	if got != want {
		t.Fatalf("buildEmailActionURL() = %q, want %q", got, want)
	}
}

func TestBuildEmailActionURLIncludesLocale(t *testing.T) {
	got := buildEmailActionURL("https://example.com/", "/activate", "token", "en-US")
	want := "https://example.com/activate?lang=en&token=token"
	if got != want {
		t.Fatalf("buildEmailActionURL() = %q, want %q", got, want)
	}
}

func TestGenerateActivationEmailBodyUsesLocale(t *testing.T) {
	body, err := generateActivationEmailBody("aki", "token", "en")
	if err != nil {
		t.Fatalf("generateActivationEmailBody() error = %v", err)
	}

	if !strings.Contains(body, "Welcome to") {
		t.Fatalf("activation email body should use English copy, got %q", body)
	}
	if !strings.Contains(body, "lang=en&amp;token=token") {
		t.Fatalf("activation email body should include lang query, got %q", body)
	}
}

func TestGeneratePasswordResetEmailBodyUsesLocale(t *testing.T) {
	body, err := generatePasswordResetEmailBody("aki", "token", "en")
	if err != nil {
		t.Fatalf("generatePasswordResetEmailBody() error = %v", err)
	}

	if !strings.Contains(body, "Password reset request") {
		t.Fatalf("password reset email body should use English copy, got %q", body)
	}
	if !strings.Contains(body, "Reset password") {
		t.Fatalf("password reset email body should include English action, got %q", body)
	}
	if !strings.Contains(body, "lang=en&amp;token=token") {
		t.Fatalf("password reset email body should include lang query, got %q", body)
	}
}

func TestNormalizeSenderUsesConfiguredNameAndEmail(t *testing.T) {
	name, email := normalizeSender(pageConfig.MailSettingsConfig{
		FromName:     " GooseForum Notice ",
		FromEmail:    " noreply@example.com ",
		SmtpUsername: "smtp@example.com",
	})

	if name != "GooseForum Notice" {
		t.Fatalf("sender name = %q, want %q", name, "GooseForum Notice")
	}
	if email != "noreply@example.com" {
		t.Fatalf("sender email = %q, want %q", email, "noreply@example.com")
	}
}

func TestNormalizeSenderFallsBackToDefaultNameAndSMTPUsername(t *testing.T) {
	name, email := normalizeSender(pageConfig.MailSettingsConfig{
		SmtpUsername: " smtp@example.com ",
	})

	if name != "GooseForum" {
		t.Fatalf("sender name = %q, want %q", name, "GooseForum")
	}
	if email != "smtp@example.com" {
		t.Fatalf("sender email = %q, want %q", email, "smtp@example.com")
	}
}

func TestNormalizeSenderUsesLocalDefaultEmailWhenMissing(t *testing.T) {
	name, email := normalizeSender(pageConfig.MailSettingsConfig{})

	if name != "GooseForum" {
		t.Fatalf("sender name = %q, want %q", name, "GooseForum")
	}
	if email != "noreply@localhost" {
		t.Fatalf("sender email = %q, want %q", email, "noreply@localhost")
	}
}
