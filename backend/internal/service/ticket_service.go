package service

import (
	"context"
	"errors"
	"strings"

	"helpdesk/backend/internal/domain"
)

var allowedSorts = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"priority":   true,
	"status":     true,
}

type TicketService struct {
	tickets  domain.TicketRepository
	comments domain.CommentRepository
	users    domain.UserRepository
}

func NewTicketService(tickets domain.TicketRepository, comments domain.CommentRepository, users domain.UserRepository) *TicketService {
	return &TicketService{tickets: tickets, comments: comments, users: users}
}

type CreateTicketInput struct {
	Title       string
	Description string
	Category    string
	Priority    string
}

type TicketDetail struct {
	Ticket   domain.Ticket
	Comments []domain.Comment
}

func (s *TicketService) Create(ctx context.Context, actor domain.User, in CreateTicketInput) (domain.Ticket, error) {
	details := map[string]string{}
	title, msg := validateTitle(in.Title)
	if msg != "" {
		details["title"] = msg
	}
	description, msg := validateDescription(in.Description)
	if msg != "" {
		details["description"] = msg
	}
	category, ok := domain.ParseCategory(in.Category)
	if !ok {
		details["category"] = "Kategori tidak valid."
	}
	priority, ok := domain.ParsePriority(in.Priority)
	if !ok {
		details["priority"] = "Prioritas tidak valid."
	}
	if len(details) > 0 {
		return domain.Ticket{}, domain.Validation(details)
	}

	ticket := domain.Ticket{
		Title:       title,
		Description: description,
		Category:    category,
		Priority:    priority,
		Status:      domain.StatusOpen,
		RequesterID: actor.ID,
	}
	return s.tickets.Create(ctx, ticket)
}

func (s *TicketService) Get(ctx context.Context, actor domain.User, id int64) (TicketDetail, error) {
	ticket, err := s.tickets.GetByID(ctx, id)
	if err != nil {
		return TicketDetail{}, err
	}
	if !canAccessTicket(actor, ticket) {
		return TicketDetail{}, domain.Forbidden("Anda tidak memiliki akses ke tiket ini.")
	}

	comments, err := s.comments.ListByTicket(ctx, id, actor.Role.IsStaff())
	if err != nil {
		return TicketDetail{}, err
	}
	return TicketDetail{Ticket: ticket, Comments: comments}, nil
}

func (s *TicketService) List(ctx context.Context, actor domain.User, filter domain.TicketFilter) ([]domain.Ticket, int, error) {
	if err := normalizeTicketFilter(&filter); err != nil {
		return nil, 0, err
	}
	// Server-enforced scope: a requester can never read other users' tickets.
	if actor.Role == domain.RoleRequester {
		requesterID := actor.ID
		filter.RequesterID = &requesterID
		filter.AssigneeID = nil
		filter.Unassigned = false
	}
	return s.tickets.List(ctx, filter)
}

func (s *TicketService) Update(ctx context.Context, actor domain.User, id int64, patch domain.TicketUpdate) (domain.Ticket, error) {
	ticket, err := s.tickets.GetByID(ctx, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !canAccessTicket(actor, ticket) {
		return domain.Ticket{}, domain.Forbidden("Anda tidak memiliki akses untuk mengubah tiket ini.")
	}
	if patch.Status == nil && patch.Priority == nil && !patch.AssigneeSet {
		return domain.Ticket{}, domain.ValidationMessage("Minimal satu field harus diubah.")
	}

	desiredStatus := ticket.Status
	if patch.Status != nil {
		if !patch.Status.Valid() {
			return domain.Ticket{}, domain.Validation(map[string]string{"status": "Status tidak valid."})
		}
		if *patch.Status != ticket.Status {
			rule, found := domain.FindTransition(ticket.Status, *patch.Status)
			if !found {
				return domain.Ticket{}, domain.InvalidTransition(ticket.Status, *patch.Status, actor.Role)
			}
			isOwner := ticket.RequesterID == actor.ID
			if !rule.Allows(actor.Role, isOwner) {
				if actor.Role == domain.RoleRequester && rule.Roles[actor.Role] && rule.RequireOwner && !isOwner {
					return domain.Ticket{}, domain.Forbidden("Anda tidak memiliki akses untuk mengubah tiket ini.")
				}
				return domain.Ticket{}, domain.InvalidTransition(ticket.Status, *patch.Status, actor.Role)
			}
			desiredStatus = *patch.Status
		}
	}

	desiredPriority := ticket.Priority
	if patch.Priority != nil {
		if !patch.Priority.Valid() {
			return domain.Ticket{}, domain.Validation(map[string]string{"priority": "Prioritas tidak valid."})
		}
		if actor.Role == domain.RoleRequester {
			return domain.Ticket{}, domain.Forbidden("Requester tidak dapat mengubah prioritas tiket.")
		}
		desiredPriority = *patch.Priority
	}

	desiredAssignee := ticket.AssigneeID
	if patch.AssigneeSet {
		if actor.Role == domain.RoleRequester {
			return domain.Ticket{}, domain.Forbidden("Requester tidak dapat mengubah assignee tiket.")
		}
		desiredAssignee = patch.AssigneeID

		switch actor.Role {
		case domain.RoleAgent:
			if patch.AssigneeID != nil && *patch.AssigneeID != actor.ID {
				return domain.Ticket{}, domain.Forbidden("Agent hanya dapat mengambil tiket untuk dirinya sendiri.")
			}
			if patch.AssigneeID != nil && ticket.AssigneeID != nil && *ticket.AssigneeID != actor.ID {
				return domain.Ticket{}, domain.Forbidden("Tiket sudah ditangani agent lain.")
			}
			if patch.AssigneeID == nil && ticket.AssigneeID != nil && *ticket.AssigneeID != actor.ID {
				return domain.Ticket{}, domain.Forbidden("Agent hanya dapat melepas tiket yang ditangani sendiri.")
			}
		case domain.RoleAdmin:
			if patch.AssigneeID != nil {
				assignee, err := s.users.GetByID(ctx, *patch.AssigneeID)
				if err != nil {
					var appErr *domain.AppError
					if errors.As(err, &appErr) && appErr.Code == "NOT_FOUND" {
						return domain.Ticket{}, domain.Validation(map[string]string{"assignee_id": "Assignee tidak ditemukan."})
					}
					return domain.Ticket{}, err
				}
				if !assignee.Role.IsStaff() {
					return domain.Ticket{}, domain.Validation(map[string]string{"assignee_id": "Assignee harus pengguna dengan peran agent atau admin."})
				}
			}
		}
	}

	// An agent taking an unassigned open ticket becomes its assignee.
	if ticket.Status != domain.StatusInProgress && desiredStatus == domain.StatusInProgress &&
		desiredAssignee == nil && actor.Role == domain.RoleAgent {
		id := actor.ID
		desiredAssignee = &id
	}

	return s.tickets.Update(ctx, id, desiredStatus, desiredPriority, desiredAssignee)
}

func canAccessTicket(actor domain.User, ticket domain.Ticket) bool {
	return actor.Role.IsStaff() || ticket.RequesterID == actor.ID
}

func normalizeTicketFilter(filter *domain.TicketFilter) error {
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Page < 1 {
		return domain.Validation(map[string]string{"page": "Parameter page tidak valid."})
	}
	if filter.PerPage == 0 {
		filter.PerPage = 20
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		return domain.Validation(map[string]string{"per_page": "per_page harus di antara 1 dan 100."})
	}
	if filter.Sort == "" {
		filter.Sort = "created_at"
	}
	if !allowedSorts[filter.Sort] {
		return domain.Validation(map[string]string{"sort": "Parameter sort tidak valid."})
	}
	if filter.Order == "" {
		filter.Order = "desc"
	}
	if filter.Order != "asc" && filter.Order != "desc" {
		return domain.Validation(map[string]string{"order": "Parameter order harus asc atau desc."})
	}
	if filter.Status != nil && !filter.Status.Valid() {
		return domain.Validation(map[string]string{"status": "Status tidak valid."})
	}
	if filter.Priority != nil && !filter.Priority.Valid() {
		return domain.Validation(map[string]string{"priority": "Prioritas tidak valid."})
	}
	if filter.AssigneeID != nil && *filter.AssigneeID <= 0 {
		return domain.Validation(map[string]string{"assignee_id": "assignee_id tidak valid."})
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if runeLen(filter.Search) > 100 {
		return domain.Validation(map[string]string{"search": "Pencarian maksimal 100 karakter."})
	}
	return nil
}
