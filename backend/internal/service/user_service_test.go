package service

import (
	"context"
	"testing"

	"helpdesk/backend/internal/domain"
)

func TestAdminCannotChangeOwnRole(t *testing.T) {
	users := newFakeUserRepo(domain.User{ID: 1, Role: domain.RoleAdmin})
	svc := NewUserService(users)

	_, err := svc.UpdateRole(context.Background(), domain.User{ID: 1, Role: domain.RoleAdmin}, 1, "agent")
	requireAppCode(t, err, "SELF_ROLE_CHANGE")
}

func TestLastAdminCannotBeDemoted(t *testing.T) {
	// The actor is an admin loaded by middleware, but only one admin remains in
	// the target data set.
	users := newFakeUserRepo(domain.User{ID: 2, Role: domain.RoleAdmin})
	svc := NewUserService(users)

	_, err := svc.UpdateRole(context.Background(), domain.User{ID: 1, Role: domain.RoleAdmin}, 2, "requester")
	requireAppCode(t, err, "LAST_ADMIN")
}

func TestAdminCanPromoteRequester(t *testing.T) {
	users := newFakeUserRepo(domain.User{ID: 3, Role: domain.RoleRequester})
	svc := NewUserService(users)

	user, err := svc.UpdateRole(context.Background(), domain.User{ID: 1, Role: domain.RoleAdmin}, 3, "agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != domain.RoleAgent {
		t.Fatalf("expected agent, got %s", user.Role)
	}
}

func TestNonAdminCannotListUsers(t *testing.T) {
	svc := NewUserService(newFakeUserRepo())

	_, _, err := svc.List(context.Background(), domain.User{ID: 5, Role: domain.RoleAgent}, domain.UserFilter{})
	requireAppCode(t, err, "FORBIDDEN")
}

func TestAdminCanListUsers(t *testing.T) {
	users := newFakeUserRepo(
		domain.User{ID: 1, Role: domain.RoleAdmin, Name: "Admin"},
		domain.User{ID: 2, Role: domain.RoleAgent, Name: "Agent"},
	)
	svc := NewUserService(users)

	list, total, err := svc.List(context.Background(), domain.User{ID: 1, Role: domain.RoleAdmin}, domain.UserFilter{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected 2 users, got total=%d list=%d", total, len(list))
	}
}
