package repository

import (
	"testing"

	"github.com/basketikun/infinite-canvas/model"
)

func TestGetUserByRegisterIPFindsExistingRegisteredUser(t *testing.T) {
	setupCreditTestDB(t)
	if _, err := SaveUser(model.User{ID: "user-ip-1", Username: "ip-user", AffCode: "aff-ip-1", RegisterIP: "203.0.113.9"}); err != nil {
		t.Fatalf("save user: %v", err)
	}
	user, ok, err := GetUserByRegisterIP("203.0.113.9")
	if err != nil {
		t.Fatalf("get by ip: %v", err)
	}
	if !ok || user.ID != "user-ip-1" {
		t.Fatalf("expected existing user by register ip, got ok=%v user=%+v", ok, user)
	}
	_, ok, err = GetUserByRegisterIP("")
	if err != nil || ok {
		t.Fatalf("empty ip should not match users, ok=%v err=%v", ok, err)
	}
}
