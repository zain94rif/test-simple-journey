package domain

import (
	"strings"
	"time"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

func (s Status) Valid() bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

func ParseStatus(raw string) (Status, bool) {
	s := Status(strings.ToLower(strings.TrimSpace(raw)))
	return s, s.Valid()
}

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func ParsePriority(raw string) (Priority, bool) {
	p := Priority(strings.ToLower(strings.TrimSpace(raw)))
	return p, p.Valid()
}

type Category string

const (
	CategoryHardware Category = "hardware"
	CategorySoftware Category = "software"
	CategoryNetwork  Category = "network"
	CategoryAccess   Category = "access"
	CategoryOther    Category = "other"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryHardware, CategorySoftware, CategoryNetwork, CategoryAccess, CategoryOther:
		return true
	default:
		return false
	}
}

func ParseCategory(raw string) (Category, bool) {
	c := Category(strings.ToLower(strings.TrimSpace(raw)))
	return c, c.Valid()
}

// TicketUser is a compact user object embedded in ticket responses.
type TicketUser struct {
	ID    int64
	Name  string
	Email string
}

type Ticket struct {
	ID          int64
	Title       string
	Description string
	Category    Category
	Priority    Priority
	Status      Status
	RequesterID int64
	AssigneeID  *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Requester *TicketUser
	Assignee  *TicketUser
}

// TicketFilter is shared between service and repository. RequesterID is the
// server-enforced scope for requester users; it is never taken from the client.
type TicketFilter struct {
	Page        int
	PerPage     int
	Status      *Status
	Priority    *Priority
	AssigneeID  *int64
	Unassigned  bool
	Search      string
	Sort        string
	Order       string
	RequesterID *int64
}

// TicketUpdate describes a PATCH. AssigneeSet distinguishes an omitted field
// from an explicit null (unassign).
type TicketUpdate struct {
	Status      *Status
	Priority    *Priority
	AssigneeID  *int64
	AssigneeSet bool
}

type Comment struct {
	ID         int64
	TicketID   int64
	AuthorID   int64
	Body       string
	IsInternal bool
	CreatedAt  time.Time
	Author     *UserSummary
}
