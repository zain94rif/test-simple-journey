package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"helpdesk/backend/internal/auth"
	"helpdesk/backend/internal/config"
	"helpdesk/backend/internal/domain"
)

type account struct {
	config config.SeedAccount
	role   domain.Role
}

// Run inserts the demo accounts if they do not exist. Existing users are never
// modified, so repeated runs are safe.
func Run(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	accounts := []account{
		{config: cfg.SeedAdmin, role: domain.RoleAdmin},
		{config: cfg.SeedAgent, role: domain.RoleAgent},
		{config: cfg.SeedRequester, role: domain.RoleRequester},
	}

	for _, acc := range accounts {
		hash, err := auth.HashPassword(acc.config.Password, cfg.BcryptCost)
		if err != nil {
			return fmt.Errorf("hash password seed %s: %w", acc.role, err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (name, email, password_hash, role)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING`,
			acc.config.Name, acc.config.Email, hash, string(acc.role),
		); err != nil {
			return fmt.Errorf("seed %s: %w", acc.role, err)
		}
	}
	return nil
}
