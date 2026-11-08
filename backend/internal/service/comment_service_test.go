package service

import (
	"context"
	"testing"

	"helpdesk/backend/internal/domain"
)

func TestRequesterCannotCreateInternalNote(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 1, Status: domain.StatusOpen})
	svc := NewCommentService(tickets, newFakeCommentRepo())

	_, err := svc.Create(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10, CreateCommentInput{
		Body:       "catatan rahasia",
		IsInternal: true,
	})
	requireAppCode(t, err, "INTERNAL_NOTE_FORBIDDEN")
}

func TestRequesterCannotCommentOnOthersTicket(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 2, Status: domain.StatusOpen})
	svc := NewCommentService(tickets, newFakeCommentRepo())

	_, err := svc.Create(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10, CreateCommentInput{
		Body: "halo",
	})
	requireAppCode(t, err, "FORBIDDEN")
}

func TestAgentCanCreateInternalNote(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 2, Status: domain.StatusOpen})
	svc := NewCommentService(tickets, newFakeCommentRepo())

	comment, err := svc.Create(context.Background(), domain.User{ID: 5, Name: "Agent Lima", Role: domain.RoleAgent}, 10, CreateCommentInput{
		Body:       "sedang dicek",
		IsInternal: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !comment.IsInternal || comment.Author == nil || comment.Author.Role != domain.RoleAgent {
		t.Fatalf("unexpected comment: %+v", comment)
	}
}

func TestCommentBodyIsRequired(t *testing.T) {
	tickets := newFakeTicketRepo(domain.Ticket{ID: 10, RequesterID: 1, Status: domain.StatusOpen})
	svc := NewCommentService(tickets, newFakeCommentRepo())

	_, err := svc.Create(context.Background(), domain.User{ID: 1, Role: domain.RoleRequester}, 10, CreateCommentInput{
		Body: "   ",
	})
	requireAppCode(t, err, "VALIDATION_ERROR")
}
