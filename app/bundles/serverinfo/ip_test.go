package serverinfo

import "testing"

func TestIp(t *testing.T) {
	localIP, localErr := GetLocalIp()
	externalIP, externalErr := ExternalIP()
	if localErr != nil && externalErr != nil {
		t.Fatalf("expected at least one IP lookup to succeed, localErr=%v externalErr=%v", localErr, externalErr)
	}
	if localErr == nil && localIP == "" {
		t.Fatal("GetLocalIp returned empty IP without error")
	}
	if externalErr == nil && externalIP == nil {
		t.Fatal("ExternalIP returned nil IP without error")
	}
}
