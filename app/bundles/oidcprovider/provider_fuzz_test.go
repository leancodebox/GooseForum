package oidcprovider

import "testing"

func FuzzPKCEValidation(f *testing.F) {
	f.Add("authorization-code-verifier-with-43-characters-minimum", "challenge")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, verifier, challenge string) {
		_ = verifyPKCE(verifier, challenge)
		_ = validPKCEValue(verifier)
	})
}

func FuzzRedirectURIValidation(f *testing.F) {
	f.Add("https://client.example/callback")
	f.Add("com.example.app:/callback")
	f.Add("javascript:alert(1)")
	f.Fuzz(func(t *testing.T, value string) { _ = validRedirectURI(value) })
}
