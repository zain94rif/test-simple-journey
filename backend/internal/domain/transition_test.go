package domain_test

import (
	"testing"

	"helpdesk/backend/internal/domain"
)

func TestTransitionMatrix(t *testing.T) {
	statuses := []domain.Status{
		domain.StatusOpen,
		domain.StatusInProgress,
		domain.StatusResolved,
		domain.StatusClosed,
	}
	roles := []domain.Role{
		domain.RoleRequester,
		domain.RoleAgent,
		domain.RoleAdmin,
	}

	// Only these from->to pairs are allowed, and only for these roles.
	allowed := map[[2]domain.Status]map[domain.Role]bool{
		{domain.StatusOpen, domain.StatusInProgress}: {
			domain.RoleAgent: true, domain.RoleAdmin: true,
		},
		{domain.StatusInProgress, domain.StatusResolved}: {
			domain.RoleAgent: true, domain.RoleAdmin: true,
		},
		{domain.StatusResolved, domain.StatusClosed}: {
			domain.RoleRequester: true, domain.RoleAgent: true, domain.RoleAdmin: true,
		},
		{domain.StatusResolved, domain.StatusOpen}: {
			domain.RoleRequester: true,
		},
		{domain.StatusOpen, domain.StatusClosed}: {
			domain.RoleRequester: true, domain.RoleAdmin: true,
		},
	}

	for _, from := range statuses {
		for _, to := range statuses {
			for _, role := range roles {
				want := allowed[[2]domain.Status{from, to}][role]
				got := domain.CanTransition(from, to, role, true)
				if got != want {
					t.Errorf("CanTransition(%s -> %s, %s) = %v, want %v", from, to, role, got, want)
				}
			}
		}
	}
}

func TestRequesterTransitionsRequireOwnership(t *testing.T) {
	cases := []struct {
		from, to domain.Status
		role     domain.Role
		owner    bool
		want     bool
	}{
		{domain.StatusOpen, domain.StatusClosed, domain.RoleRequester, false, false},
		{domain.StatusResolved, domain.StatusClosed, domain.RoleRequester, false, false},
		{domain.StatusResolved, domain.StatusOpen, domain.RoleRequester, false, false},
		{domain.StatusOpen, domain.StatusClosed, domain.RoleRequester, true, true},
		{domain.StatusResolved, domain.StatusClosed, domain.RoleAgent, false, true},
		{domain.StatusOpen, domain.StatusClosed, domain.RoleAdmin, false, true},
	}
	for _, tc := range cases {
		got := domain.CanTransition(tc.from, tc.to, tc.role, tc.owner)
		if got != tc.want {
			t.Errorf("CanTransition(%s -> %s, %s, owner=%v) = %v, want %v",
				tc.from, tc.to, tc.role, tc.owner, got, tc.want)
		}
	}
}

func TestClosedIsFinal(t *testing.T) {
	for _, role := range []domain.Role{domain.RoleRequester, domain.RoleAgent, domain.RoleAdmin} {
		for _, to := range []domain.Status{domain.StatusOpen, domain.StatusInProgress, domain.StatusResolved} {
			if domain.CanTransition(domain.StatusClosed, to, role, true) {
				t.Errorf("closed -> %s should not be allowed for %s", to, role)
			}
		}
	}
}
