package pageConfig

import "testing"

func TestSecuritySettingsDefaultsOnlyApplyToAbsentFields(t *testing.T) {
	defaults := SecurityAndRegistration{EnableSignup: true, RegistrationIPLimit: 5, RegistrationEmailLimit: 3, RegistrationGlobalLimit: 100}
	old := DecodeSecuritySettings(`{"enableSignup":false,"allowedDomains":["example.com"]}`, defaults)
	if old.EnableSignup || old.RegistrationIPLimit != 5 || old.RegistrationEmailLimit != 3 || old.RegistrationGlobalLimit != 100 {
		t.Fatalf("old config: %+v", old)
	}
	disabled := DecodeSecuritySettings(`{"registrationIPLimit":0,"registrationEmailLimit":0,"registrationGlobalLimit":0}`, defaults)
	if disabled.RegistrationIPLimit != 0 || disabled.RegistrationEmailLimit != 0 || disabled.RegistrationGlobalLimit != 0 {
		t.Fatalf("disabled config: %+v", disabled)
	}
}
