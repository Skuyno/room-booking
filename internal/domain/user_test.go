package domain

import "testing"

func TestRoleMappings(t *testing.T) {
	t.Parallel()

	if RoleCodeByID(RoleIDAdmin) != RoleAdmin {
		t.Fatalf("RoleCodeByID(admin) = %q", RoleCodeByID(RoleIDAdmin))
	}
	if RoleCodeByID(RoleIDUser) != RoleUser {
		t.Fatalf("RoleCodeByID(user) = %q", RoleCodeByID(RoleIDUser))
	}
	if RoleCodeByID(99) != "" {
		t.Fatalf("RoleCodeByID(99) = %q, want empty string", RoleCodeByID(99))
	}

	if RoleIDByCode(RoleAdmin) != RoleIDAdmin {
		t.Fatalf("RoleIDByCode(admin) = %d", RoleIDByCode(RoleAdmin))
	}
	if RoleIDByCode(RoleUser) != RoleIDUser {
		t.Fatalf("RoleIDByCode(user) = %d", RoleIDByCode(RoleUser))
	}
	if RoleIDByCode("guest") != 0 {
		t.Fatalf("RoleIDByCode(guest) = %d, want 0", RoleIDByCode("guest"))
	}
}
