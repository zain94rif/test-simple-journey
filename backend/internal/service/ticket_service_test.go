package service

import (
	"context"
	"errors"
	"testing"

	"helpdesk/backend/internal/domain"
)

func requireAppCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError %s, got %T: %v", code, err, err)
	}
	if appErr.Code != code {
		t.Fatalf("expected code %s, got %s (%s)", code, appErr.Code, appErr.Message)
	}
}

func pointer[T any](v T) *T { return &v }

func TestRequesterCannotReadOthersTicket(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 2, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	_, err := svc.Get(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10)
	requireAppCode(t, err, "FORBIDDEN")
}

func TestRequesterListIsScopedToOwnTickets(t *testing.T) {
	tickets := newFakeTicketRepo(
		domain.Ticket{ID: 1, RequesterID: 1, Title: "Milik saya", Status: domain.StatusOpen},
		domain.Ticket{ID: 2, RequesterID: 2, Title: "Milik orang lain", Status: domain.StatusOpen},
	)
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	got, total, err := svc.List(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, domain.TicketFilter{
		Page:       1,
		PerPage:    20,
		AssigneeID: pointer(int64(99)), // must be ignored for requesters
		Unassigned: true,               // must be ignored for requesters
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("expected only ticket 1, got total=%d tickets=%+v", total, got)
	}
}

func TestRequesterCannotSeeInternalComments(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 1, Status: domain.StatusInProgress})
	comments := newFakeCommentRepo(
		domain.Comment{ID: 1, TicketID: 10, Body: "publik", IsInternal: false},
		domain.Comment{ID: 2, TicketID: 10, Body: "internal", IsInternal: true},
	)
	svc := NewTicketService(tickets, comments, newFakeUserRepo())

	detail, err := svc.Get(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detail.Comments) != 1 || detail.Comments[0].Body != "publik" {
		t.Fatalf("expected only public comment, got %+v", detail.Comments)
	}
}

func TestAgentClaimAutoAssignsSelf(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 3, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	ticket, err := svc.Update(context.Background(), domain.User{ID: 5, Role: domain.RoleAgent}, 10, domain.TicketUpdate{
		Status: pointer(domain.StatusInProgress),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Status != domain.StatusInProgress {
		t.Fatalf("expected in_progress, got %s", ticket.Status)
	}
	if ticket.AssigneeID == nil || *ticket.AssigneeID != 5 {
		t.Fatalf("expected assignee 5, got %v", ticket.AssigneeID)
	}
}

func TestAgentCannotAssignAnotherAgent(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 3, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	_, err := svc.Update(context.Background(), domain.User{ID: 5, Role: domain.RoleAgent}, 10, domain.TicketUpdate{
		AssigneeID:  pointer(int64(6)),
		AssigneeSet: true,
	})
	requireAppCode(t, err, "FORBIDDEN")
}

func TestAgentCannotTakeOverAnotherAgentsTicket(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{
		ID: 10, RequesterID: 3, Status: domain.StatusOpen, AssigneeID: pointer(int64(6)),
	})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	_, err := svc.Update(context.Background(), domain.User{ID: 5, Role: domain.RoleAgent}, 10, domain.TicketUpdate{
		AssigneeID:  pointer(int64(5)),
		AssigneeSet: true,
	})
	requireAppCode(t, err, "FORBIDDEN")
}

func TestInvalidTransitionIsRejected(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 3, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	_, err := svc.Update(context.Background(), domain.User{ID: 5, Role: domain.RoleAgent}, 10, domain.TicketUpdate{
		Status: pointer(domain.StatusResolved),
	})
	requireAppCode(t, err, "INVALID_TRANSITION")
}

func TestCreateTicketForcesOpenStatusAndActorAsRequester(t *testing.T) {
	tickets := newFakeTicketRepo()
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	ticket, err := svc.Create(context.Background(), domain.User{ID: 7, Role: domain.RoleRequester}, CreateTicketInput{
		Title:       "Printer macet",
		Description: "Printer tidak bisa mencetak dokumen apapun.",
		Category:    "hardware",
		Priority:    "high",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Status != domain.StatusOpen || ticket.RequesterID != 7 || ticket.AssigneeID != nil {
		t.Fatalf("unexpected ticket defaults: %+v", ticket)
	}
}

func TestRequesterCannotChangePriority(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 1, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	_, err := svc.Update(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10, domain.TicketUpdate{
		Priority: pointer(domain.PriorityUrgent),
	})
	requireAppCode(t, err, "FORBIDDEN")
}

func TestRequesterCanCloseOwnOpenTicket(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 1, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	ticket, err := svc.Update(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10, domain.TicketUpdate{
		Status: pointer(domain.StatusClosed),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.Status != domain.StatusClosed {
		t.Fatalf("expected closed, got %s", ticket.Status)
	}
}

func TestRequesterCannotCloseAnotherUsersTicket(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 2, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), newFakeUserRepo())

	_, err := svc.Update(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10, domain.TicketUpdate{
		Status: pointer(domain.StatusClosed),
	})
	requireAppCode(t, err, "FORBIDDEN")
}

func TestAdminCanAssignAnyAgent(t *testing.T) {
	users := newFakeUserRepo(
		domain.User{ID: 1, Role: domain.RoleAdmin},
		domain.User{ID: 5, Role: domain.RoleAgent},
	)
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 3, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), users)

	ticket, err := svc.Update(context.Background(), domain.User{ID: 1, Role: domain.RoleAdmin}, 10, domain.TicketUpdate{
		AssigneeID:  pointer(int64(5)),
		AssigneeSet: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ticket.AssigneeID == nil || *ticket.AssigneeID != 5 {
		t.Fatalf("expected assignee 5, got %v", ticket.AssigneeID)
	}
}

func TestAdminCannotAssignRequester(t *testing.T) {
	users := newFakeUserRepo(
		domain.User{ID: 1, Role: domain.RoleAdmin},
		domain.User{ID: 3, Role: domain.RoleRequester},
	)
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 3, Status: domain.StatusOpen})
	svc := NewTicketService(tickets, newFakeCommentRepo(), users)

	_, err := svc.Update(context.Background(), domain.User{ID: 1, Role: domain.RoleAdmin}, 10, domain.TicketUpdate{
		AssigneeID:  pointer(int64(3)),
		AssigneeSet: true,
	})
	requireAppCode(t, err, "VALIDATION_ERROR")
}
