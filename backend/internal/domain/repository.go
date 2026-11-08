package domain

import "context"

type UserFilter struct {
	Page    int
	PerPage int
	Role    *Role
	Search  string
}

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByID(ctx context.Context, id int64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	List(ctx context.Context, filter UserFilter) ([]User, int, error)
	UpdateRole(ctx context.Context, id int64, role Role) (User, error)
	CountByRole(ctx context.Context, role Role) (int, error)
}

type TicketRepository interface {
	Create(ctx context.Context, ticket Ticket) (Ticket, error)
	GetByID(ctx context.Context, id int64) (Ticket, error)
	List(ctx context.Context, filter TicketFilter) ([]Ticket, int, error)
	Update(ctx context.Context, id int64, status Status, priority Priority, assigneeID *int64) (Ticket, error)
	Touch(ctx context.Context, id int64) error
}

type CommentRepository interface {
	Create(ctx context.Context, comment Comment) (Comment, error)
	ListByTicket(ctx context.Context, ticketID int64, includeInternal bool) ([]Comment, error)
}
