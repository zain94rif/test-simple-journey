package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"helpdesk/backend/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, role, created_at`,
		user.Name, user.Email, user.PasswordHash, string(user.Role),
	)
	created, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.DuplicateEmail()
		}
		return domain.User{}, fmt.Errorf("buat user: %w", err)
	}
	return created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE id = $1`, id)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.NotFound("Pengguna")
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("ambil user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE email = lower($1)`, email)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.NotFound("Pengguna")
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("ambil user berdasarkan email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 4)
	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}

	if filter.Role != nil {
		add("role = $%d", string(*filter.Role))
	}
	if filter.Search != "" {
		args = append(args, filter.Search)
		n := len(args)
		where = append(where, fmt.Sprintf(
			"(name ILIKE '%%' || $%d || '%%' OR email ILIKE '%%' || $%d || '%%')", n, n,
		))
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hitung user: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	args = append(args, perPage, (page-1)*perPage)
	query := fmt.Sprintf(`
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`, whereSQL, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("daftar user: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterasi user: %w", err)
	}
	return users, total, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role domain.Role) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users SET role = $2 WHERE id = $1
		RETURNING id, name, email, password_hash, role, created_at`, id, string(role))
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.NotFound("Pengguna")
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("ubah peran user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) CountByRole(ctx context.Context, role domain.Role) (int, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE role = $1`, string(role)).Scan(&count); err != nil {
		return 0, fmt.Errorf("hitung user per peran: %w", err)
	}
	return count, nil
}

func scanUser(row pgx.Row) (domain.User, error) {
	var (
		user domain.User
		role string
	)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &role, &user.CreatedAt); err != nil {
		return domain.User{}, err
	}
	user.Role = domain.Role(role)
	return user, nil
}
