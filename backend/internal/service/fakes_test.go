package service

import (
	"context"
	"strings"
	"time"

	"helpdesk/backend/internal/domain"
)

type fakeUserRepo struct {
	users  map[int64]domain.User
	nextID int64
}

func newFakeUserRepo(users ...domain.User) *fakeUserRepo {
	repo := &fakeUserRepo{users: make(map[int64]domain.User), nextID: 1}
	for _, user := range users {
		repo.users[user.ID] = user
		if user.ID >= repo.nextID {
			repo.nextID = user.ID + 1
		}
	}
	return repo
}

func (f *fakeUserRepo) Create(_ context.Context, user domain.User) (domain.User, error) {
	for _, existing := range f.users {
		if strings.EqualFold(existing.Email, user.Email) {
			return domain.User{}, domain.DuplicateEmail()
		}
	}
	user.ID = f.nextID
	f.nextID++
	user.CreatedAt = time.Now().UTC()
	f.users[user.ID] = user
	return user, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id int64) (domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return domain.User{}, domain.NotFound("Pengguna")
	}
	return user, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	for _, user := range f.users {
		if strings.EqualFold(user.Email, email) {
			return user, nil
		}
	}
	return domain.User{}, domain.NotFound("Pengguna")
}

func (f *fakeUserRepo) List(_ context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	out := make([]domain.User, 0)
	for _, user := range f.users {
		if filter.Role != nil && user.Role != *filter.Role {
			continue
		}
		if filter.Search != "" {
			q := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(user.Name), q) && !strings.Contains(strings.ToLower(user.Email), q) {
				continue
			}
		}
		out = append(out, user)
	}
	return out, len(out), nil
}

func (f *fakeUserRepo) UpdateRole(_ context.Context, id int64, role domain.Role) (domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return domain.User{}, domain.NotFound("Pengguna")
	}
	user.Role = role
	f.users[id] = user
	return user, nil
}

func (f *fakeUserRepo) CountByRole(_ context.Context, role domain.Role) (int, error) {
	count := 0
	for _, user := range f.users {
		if user.Role == role {
			count++
		}
	}
	return count, nil
}

type fakeTicketRepo struct {
	tickets map[int64]domain.Ticket
	nextID  int64
}

func newFakeTicketRepo(tickets ...domain.Ticket) *fakeTicketRepo {
	repo := &fakeTicketRepo{tickets: make(map[int64]domain.Ticket), nextID: 1}
	for _, ticket := range tickets {
		repo.tickets[ticket.ID] = ticket
		if ticket.ID >= repo.nextID {
			repo.nextID = ticket.ID + 1
		}
	}
	return repo
}

func (f *fakeTicketRepo) Create(_ context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	ticket.ID = f.nextID
	f.nextID++
	now := time.Now().UTC()
	ticket.CreatedAt = now
	ticket.UpdatedAt = now
	f.tickets[ticket.ID] = ticket
	return ticket, nil
}

func (f *fakeTicketRepo) GetByID(_ context.Context, id int64) (domain.Ticket, error) {
	ticket, ok := f.tickets[id]
	if !ok {
		return domain.Ticket{}, domain.NotFound("Tiket")
	}
	return ticket, nil
}

func (f *fakeTicketRepo) List(_ context.Context, filter domain.TicketFilter) ([]domain.Ticket, int, error) {
	out := make([]domain.Ticket, 0)
	for _, ticket := range f.tickets {
		if filter.RequesterID != nil && ticket.RequesterID != *filter.RequesterID {
			continue
		}
		if filter.Status != nil && ticket.Status != *filter.Status {
			continue
		}
		if filter.Priority != nil && ticket.Priority != *filter.Priority {
			continue
		}
		if filter.AssigneeID != nil {
			if ticket.AssigneeID == nil || *ticket.AssigneeID != *filter.AssigneeID {
				continue
			}
		}
		if filter.Unassigned && ticket.AssigneeID != nil {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(ticket.Title), strings.ToLower(filter.Search)) {
			continue
		}
		out = append(out, ticket)
	}
	return out, len(out), nil
}

func (f *fakeTicketRepo) Update(_ context.Context, id int64, status domain.Status, priority domain.Priority, assigneeID *int64) (domain.Ticket, error) {
	ticket, ok := f.tickets[id]
	if !ok {
		return domain.Ticket{}, domain.NotFound("Tiket")
	}
	ticket.Status = status
	ticket.Priority = priority
	ticket.AssigneeID = assigneeID
	ticket.UpdatedAt = time.Now().UTC()
	f.tickets[id] = ticket
	return ticket, nil
}

func (f *fakeTicketRepo) Touch(_ context.Context, id int64) error {
	ticket, ok := f.tickets[id]
	if !ok {
		return domain.NotFound("Tiket")
	}
	ticket.UpdatedAt = time.Now().UTC()
	f.tickets[id] = ticket
	return nil
}

type fakeCommentRepo struct {
	comments []domain.Comment
	nextID   int64
}

func newFakeCommentRepo(comments ...domain.Comment) *fakeCommentRepo {
	repo := &fakeCommentRepo{comments: comments, nextID: int64(len(comments) + 1)}
	return repo
}

func (f *fakeCommentRepo) Create(_ context.Context, comment domain.Comment) (domain.Comment, error) {
	comment.ID = f.nextID
	f.nextID++
	comment.CreatedAt = time.Now().UTC()
	f.comments = append(f.comments, comment)
	return comment, nil
}

func (f *fakeCommentRepo) ListByTicket(_ context.Context, ticketID int64, includeInternal bool) ([]domain.Comment, error) {
	out := make([]domain.Comment, 0)
	for _, comment := range f.comments {
		if comment.TicketID != ticketID {
			continue
		}
		if !includeInternal && comment.IsInternal {
			continue
		}
		out = append(out, comment)
	}
	return out, nil
}
