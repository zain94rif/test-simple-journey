package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"helpdesk/backend/migrations"
)

const advisoryLockID int64 = 827413001

// Run applies all pending .up.sql migrations exactly once. It is safe to call
// on every startup: a schema_migrations table tracks applied versions and an
// advisory lock prevents concurrent runners from racing.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("buat schema_migrations: %w", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("ambil koneksi migrasi: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockID); err != nil {
		return fmt.Errorf("ambil advisory lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, advisoryLockID)
	}()

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("baca folder migrasi: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		version, err := parseVersion(entry.Name())
		if err != nil {
			return err
		}

		var applied bool
		if err := conn.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("cek migrasi %d: %w", version, err)
		}
		if applied {
			continue
		}

		sqlBytes, err := migrations.FS.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("baca %s: %w", entry.Name(), err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("mulai transaksi migrasi %d: %w", version, err)
		}
		// PgConn().Exec uses the simple protocol and supports multiple
		// statements in one file while staying inside the transaction.
		if _, err := tx.Conn().PgConn().Exec(ctx, string(sqlBytes)).ReadAll(); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("jalankan migrasi %s: %w", entry.Name(), err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, version,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("catat migrasi %d: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migrasi %d: %w", version, err)
		}
	}
	return nil
}

func parseVersion(name string) (int64, error) {
	part, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0, fmt.Errorf("nama migrasi tidak valid: %s", name)
	}
	version, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("versi migrasi tidak valid pada %s: %w", name, err)
	}
	return version, nil
}
