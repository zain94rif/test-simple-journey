package service

import (
	"context"

	"helpdesk/backend/internal/domain"
)

type CommentService struct {
	tickets  domain.TicketRepository
	comments domain.CommentRepository
}

func NewCommentService(tickets domain.TicketRepository, comments domain.CommentRepository) *CommentService {
	return &CommentService{tickets: tickets, comments: comments}
}

type CreateCommentInput struct {
	Body       string
	IsInternal bool
}

func (s *CommentService) Create(ctx context.Context, actor domain.User, ticketID int64, in CreateCommentInput) (domain.Comment, error) {
	ticket, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return domain.Comment{}, err
	}
	if !canAccessTicket(actor, ticket) {
		return domain.Comment{}, domain.Forbidden("Anda tidak memiliki akses ke tiket ini.")
	}

	body, msg := validateCommentBody(in.Body)
	if msg != "" {
		return domain.Comment{}, domain.Validation(map[string]string{"body": msg})
	}
	if in.IsInternal && actor.Role == domain.RoleRequester {
		return domain.Comment{}, domain.InternalNoteForbidden()
	}

	comment, err := s.comments.Create(ctx, domain.Comment{
		TicketID:   ticketID,
		AuthorID:   actor.ID,
		Body:       body,
		IsInternal: in.IsInternal,
	})
	if err != nil {
		return domain.Comment{}, err
	}
	comment.Author = &domain.UserSummary{
		ID:    actor.ID,
		Name:  actor.Name,
		Email: actor.Email,
		Role:  actor.Role,
	}

	// Keep the ticket's updated_at meaningful without failing the request if
	// this secondary write fails.
	_ = s.tickets.Touch(ctx, ticketID)
	return comment, nil
}
