package domain

// TransitionRule describes one allowed status transition.
type TransitionRule struct {
	From         Status
	To           Status
	Roles        map[Role]bool
	RequireOwner bool
}

// Allows reports whether role may perform the transition. RequireOwner only
// applies to requesters; staff can act on any ticket.
func (r TransitionRule) Allows(role Role, isOwner bool) bool {
	if !r.Roles[role] {
		return false
	}
	if role == RoleRequester && r.RequireOwner && !isOwner {
		return false
	}
	return true
}

// transitionRules is the single source of truth from PRD-00 §6.2.
var transitionRules = []TransitionRule{
	{
		From:  StatusOpen,
		To:    StatusInProgress,
		Roles: map[Role]bool{RoleAgent: true, RoleAdmin: true},
	},
	{
		From:  StatusInProgress,
		To:    StatusResolved,
		Roles: map[Role]bool{RoleAgent: true, RoleAdmin: true},
	},
	{
		From:         StatusResolved,
		To:           StatusClosed,
		Roles:        map[Role]bool{RoleRequester: true, RoleAgent: true, RoleAdmin: true},
		RequireOwner: true,
	},
	{
		From:         StatusResolved,
		To:           StatusOpen,
		Roles:        map[Role]bool{RoleRequester: true},
		RequireOwner: true,
	},
	{
		From:         StatusOpen,
		To:           StatusClosed,
		Roles:        map[Role]bool{RoleRequester: true, RoleAdmin: true},
		RequireOwner: true,
	},
}

// FindTransition returns the rule for from -> to.
func FindTransition(from, to Status) (TransitionRule, bool) {
	for _, rule := range transitionRules {
		if rule.From == from && rule.To == to {
			return rule, true
		}
	}
	return TransitionRule{}, false
}

// CanTransition is a convenience wrapper used by tests and the service.
func CanTransition(from, to Status, role Role, isOwner bool) bool {
	rule, ok := FindTransition(from, to)
	if !ok {
		return false
	}
	return rule.Allows(role, isOwner)
}

// TransitionRules returns a copy of the official matrix (for tests/docs).
func TransitionRules() []TransitionRule {
	out := make([]TransitionRule, len(transitionRules))
	copy(out, transitionRules)
	return out
}
