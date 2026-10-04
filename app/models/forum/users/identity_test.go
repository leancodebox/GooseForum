package users

import "testing"

func TestEmailVerificationPreservesLegacyAndEnforcesNewRequirement(t *testing.T) {
	legacy := EntityComplete{IsActivated: ActivationPending}
	if legacy.NeedsEmailVerification(false) {
		t.Fatal("legacy account locked when email verification is off")
	}
	if !legacy.NeedsEmailVerification(true) {
		t.Fatal("enabled policy ignored")
	}
	legacy.RequiresEmailVerification = true
	if !legacy.NeedsEmailVerification(false) {
		t.Fatal("new account bypassed mandatory verification")
	}
	legacy.IsActivated = ActivationSuccess
	if legacy.NeedsEmailVerification(true) {
		t.Fatal("verified account was blocked")
	}
}
