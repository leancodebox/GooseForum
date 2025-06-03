package algorithm

import "testing"

func TestGenerateSigningKey(t *testing.T) {
	a, err := GenerateRandomBytes(32)
	if err != nil {
		t.Error(err)
	}
	t.Log(a)

}
