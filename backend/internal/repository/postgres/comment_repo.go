package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"helpdesk/backend/internal/domain"
)

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

func (r *CommentRepository) Create(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO comments (ticket_id, author_id, body, is_internal)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		comment.TicketID, comment.AuthorID, comment.Body, comment.IsInternal,
	).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("buat komentar: %w", err)
	}
	return comment, nil
}

func (r *CommentRepository) ListByTicket(ctx context.Context, ticketID int64, includeInternal bool) ([]domain.Comment, error) {
	query := `
		SELECT c.id, c.ticket_id, c.author_id, c.body, c.is_internal, c.created_at,
		       u.id, u.name, u.role
		FROM comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.ticket_id = $1`
	if !includeInternal {
		query += " AND c.is_internal = false"
	}
	query += " ORDER BY c.created_at ASC, c.id ASC"

	rows, err := r.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("daftar komentar: %w", err)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var (
			comment domain.Comment
			author  domain.UserSummary
			role    string
		)
		if err := rows.Scan(
			&comment.ID, &comment.TicketID, &comment.AuthorID, &comment.Body,
			&comment.IsInternal, &comment.CreatedAt,
			&author.ID, &author.Name, &role,
		); err != nil {
			return nil, fmt.Errorf("scan komentar: %w", err)
		}
		author.Role = domain.Role(role)
		comment.Author = &author
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi komentar: %w", err)
	}
	return comments, nil
}
