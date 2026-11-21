package service

import "testing"

func TestRoleValid(t *testing.T) {
	for _, role := range []string{"user", "admin", "apoteker"} {
		if !roleValid(role) {
			t.Fatalf("role %q harus valid", role)
		}
	}
	if roleValid("pemilik") {
		t.Fatal("role tidak dikenal harus ditolak")
	}
}
