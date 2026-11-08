package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"helpdesk/backend/internal/domain"
)

const ticketSelect = `
	SELECT t.id, t.title, t.description, t.category, t.priority, t.status,
	       t.requester_id, t.assignee_id, t.created_at, t.updated_at,
	       ru.id, ru.name, ru.email,
	       au.id, au.name, au.email
	FROM tickets t
	JOIN users ru ON ru.id = t.requester_id
	LEFT JOIN users au ON au.id = t.assignee_id`

var ticketSortExpressions = map[string]string{
	"created_at": "t.created_at",
	"updated_at": "t.updated_at",
	"priority": `CASE t.priority
		WHEN 'urgent' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END`,
	"status": `CASE t.status
		WHEN 'open' THEN 1 WHEN 'in_progress' THEN 2 WHEN 'resolved' THEN 3 ELSE 4 END`,
}

type TicketRepository struct {
	pool *pgxpool.Pool
}

func NewTicketRepository(pool *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{pool: pool}
}

func (r *TicketRepository) Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tickets (title, description, category, priority, status, requester_id, assignee_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		ticket.Title, ticket.Description, string(ticket.Category), string(ticket.Priority),
		string(ticket.Status), ticket.RequesterID, ticket.AssigneeID,
	).Scan(&id)
	if err != nil {
		return domain.Ticket{}, fmt.Errorf("buat tiket: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *TicketRepository) GetByID(ctx context.Context, id int64) (domain.Ticket, error) {
	row := r.pool.QueryRow(ctx, ticketSelect+" WHERE t.id = $1", id)
	ticket, err := scanTicket(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Ticket{}, domain.NotFound("Tiket")
	}
	if err != nil {
		return domain.Ticket{}, fmt.Errorf("ambil tiket: %w", err)
	}
	return ticket, nil
}

func (r *TicketRepository) List(ctx context.Context, filter domain.TicketFilter) ([]domain.Ticket, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 8)
	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}

	if filter.RequesterID != nil {
		add("t.requester_id = $%d", *filter.RequesterID)
	}
	if filter.Status != nil {
		add("t.status = $%d", string(*filter.Status))
	}
	if filter.Priority != nil {
		add("t.priority = $%d", string(*filter.Priority))
	}
	if filter.AssigneeID != nil {
		add("t.assignee_id = $%d", *filter.AssigneeID)
	} else if filter.Unassigned {
		where = append(where, "t.assignee_id IS NULL")
	}
	if filter.Search != "" {
		args = append(args, filter.Search)
		where = append(where, fmt.Sprintf("t.title ILIKE '%%' || $%d || '%%'", len(args)))
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM tickets t WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hitung tiket: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	sortExpr, ok := ticketSortExpressions[filter.Sort]
	if !ok {
		sortExpr = "t.created_at"
	}
	direction := "DESC"
	if strings.EqualFold(filter.Order, "asc") {
		direction = "ASC"
	}

	args = append(args, perPage, (page-1)*perPage)
	query := fmt.Sprintf(`%s WHERE %s ORDER BY %s %s, t.id DESC LIMIT $%d OFFSET $%d`,
		ticketSelect, whereSQL, sortExpr, direction, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("daftar tiket: %w", err)
	}
	defer rows.Close()

	tickets := make([]domain.Ticket, 0)
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan tiket: %w", err)
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterasi tiket: %w", err)
	}
	return tickets, total, nil
}

func (r *TicketRepository) Update(ctx context.Context, id int64, status domain.Status, priority domain.Priority, assigneeID *int64) (domain.Ticket, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE tickets
		SET status = $2, priority = $3, assignee_id = $4, updated_at = now()
		WHERE id = $1`,
		id, string(status), string(priority), assigneeID,
	)
	if err != nil {
		return domain.Ticket{}, fmt.Errorf("ubah tiket: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Ticket{}, domain.NotFound("Tiket")
	}
	return r.GetByID(ctx, id)
}

func (r *TicketRepository) Touch(ctx context.Context, id int64) error {
	if _, err := r.pool.Exec(ctx, `UPDATE tickets SET updated_at = now() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("perbarui updated_at tiket: %w", err)
	}
	return nil
}

func scanTicket(row pgx.Row) (domain.Ticket, error) {
	var (
		ticket                        domain.Ticket
		category, priority, status    string
		assigneeID, assigneeUserID    sql.NullInt64
		assigneeName, assigneeEmail   sql.NullString
		requesterID                   int64
		requesterName, requesterEmail string
	)
	if err := row.Scan(
		&ticket.ID, &ticket.Title, &ticket.Description, &category, &priority, &status,
		&ticket.RequesterID, &assigneeID, &ticket.CreatedAt, &ticket.UpdatedAt,
		&requesterID, &requesterName, &requesterEmail,
		&assigneeUserID, &assigneeName, &assigneeEmail,
	); err != nil {
		return domain.Ticket{}, err
	}

	ticket.Category = domain.Category(category)
	ticket.Priority = domain.Priority(priority)
	ticket.Status = domain.Status(status)
	ticket.Requester = &domain.TicketUser{ID: requesterID, Name: requesterName, Email: requesterEmail}

	if assigneeID.Valid {
		id := assigneeID.Int64
		ticket.AssigneeID = &id
	}
	if assigneeUserID.Valid {
		ticket.Assignee = &domain.TicketUser{
			ID:    assigneeUserID.Int64,
			Name:  assigneeName.String,
			Email: assigneeEmail.String,
		}
	}
	return ticket, nil
}
