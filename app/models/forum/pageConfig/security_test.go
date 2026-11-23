package pageConfig

import "testing"

func TestSecuritySettingsDefaultsOnlyApplyToAbsentFields(t *testing.T) {
	defaults := SecurityAndRegistration{EnableSignup: true, RegistrationIPLimit: 5, RegistrationEmailLimit: 3, RegistrationGlobalLimit: 100}
	old, err := decodeSecuritySettings(`{"enableSignup":false,"allowedDomains":["example.com"]}`, defaults)
	if err != nil {
		t.Fatal(err)
	}
	if old.EnableSignup || old.RegistrationIPLimit != 5 || old.RegistrationEmailLimit != 3 || old.RegistrationGlobalLimit != 100 {
		t.Fatalf("old config: %+v", old)
	}
	disabled, err := decodeSecuritySettings(`{"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`, defaults)
	if err != nil {
		t.Fatal(err)
	}
	if disabled.RegistrationIPLimit != 0 || disabled.RegistrationEmailLimit != 0 || disabled.RegistrationGlobalLimit != 0 {
		t.Fatalf("disabled config: %+v", disabled)
	}
}

func TestInvalidSecuritySettings(t *testing.T) {
	for _, raw := range []string{"null", `{"enableSignup":false,"registrationIPLimit":"oops"}`} {
		if _, err := decodeSecuritySettings(raw, SecurityAndRegistration{}); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
