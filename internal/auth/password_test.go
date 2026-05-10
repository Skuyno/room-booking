package auth

import "testing"

func TestPasswordHashAndCheck(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" || hash == "secret-pass" {
		t.Fatalf("hash = %q, expected hashed value", hash)
	}
	if err = CheckPassword(hash, "secret-pass"); err != nil {
		t.Fatalf("CheckPassword() error = %v", err)
	}
	if err = CheckPassword(hash, "wrong-pass"); err == nil {
		t.Fatal("CheckPassword() expected error for wrong password")
	}
}
