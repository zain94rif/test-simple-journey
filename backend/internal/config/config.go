package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SeedAccount describes one demo account created by the seeder.
type SeedAccount struct {
	Name     string
	Email    string
	Password string
}

// Config is the complete application configuration, loaded once at startup.
type Config struct {
	AppEnv        string
	Port          string
	DatabaseURL   string
	DBMaxConns    int32
	JWTSecret     string
	JWTTTL        time.Duration
	BcryptCost    int
	CORSOrigins   []string
	AutoMigrate   bool
	AutoSeed      bool
	SeedAdmin     SeedAccount
	SeedAgent     SeedAccount
	SeedRequester SeedAccount
	LogLevel      string
	MaxBodyBytes  int64
}

// Load reads configuration from environment variables and validates it.
func Load() (Config, error) {
	var err error
	cfg := Config{
		AppEnv:      getenv("APP_ENV", "development"),
		Port:        getenv("APP_PORT", "8080"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}

	if cfg.DBMaxConns, err = parseInt32("DB_MAX_CONNS", 10); err != nil {
		return Config{}, err
	}
	if cfg.JWTTTL, err = parseDuration("JWT_TTL", 24*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.BcryptCost, err = parseInt("BCRYPT_COST", 12); err != nil {
		return Config{}, err
	}
	if cfg.AutoMigrate, err = parseBool("AUTO_MIGRATE", true); err != nil {
		return Config{}, err
	}
	if cfg.AutoSeed, err = parseBool("AUTO_SEED", true); err != nil {
		return Config{}, err
	}
	if cfg.MaxBodyBytes, err = parseInt64("MAX_BODY_BYTES", 1<<20); err != nil {
		return Config{}, err
	}

	cfg.CORSOrigins = splitCSV(getenv("CORS_ORIGINS", "http://localhost:5173"))

	cfg.SeedAdmin = SeedAccount{
		Name:     getenv("SEED_ADMIN_NAME", "Admin Helpdesk"),
		Email:    strings.ToLower(getenv("SEED_ADMIN_EMAIL", "admin@helpdesk.local")),
		Password: os.Getenv("SEED_ADMIN_PASSWORD"),
	}
	cfg.SeedAgent = SeedAccount{
		Name:     getenv("SEED_AGENT_NAME", "Agent Satu"),
		Email:    strings.ToLower(getenv("SEED_AGENT_EMAIL", "agent@helpdesk.local")),
		Password: os.Getenv("SEED_AGENT_PASSWORD"),
	}
	cfg.SeedRequester = SeedAccount{
		Name:     getenv("SEED_REQUESTER_NAME", "Requester Satu"),
		Email:    strings.ToLower(getenv("SEED_REQUESTER_EMAIL", "requester@helpdesk.local")),
		Password: os.Getenv("SEED_REQUESTER_PASSWORD"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL wajib diisi")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET wajib diisi dan minimal 32 karakter")
	}
	if c.BcryptCost < 10 || c.BcryptCost > 14 {
		return fmt.Errorf("BCRYPT_COST harus di antara 10 dan 14")
	}
	if c.MaxBodyBytes <= 0 {
		return fmt.Errorf("MAX_BODY_BYTES harus lebih besar dari 0")
	}
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("APP_PORT tidak valid")
	}
	if c.AutoSeed {
		for label, acc := range map[string]SeedAccount{
			"admin":     c.SeedAdmin,
			"agent":     c.SeedAgent,
			"requester": c.SeedRequester,
		} {
			if acc.Name == "" || acc.Email == "" {
				return fmt.Errorf("seed %s: nama dan email wajib diisi", label)
			}
			if acc.Password == "" {
				return fmt.Errorf("seed %s: SEED_%s_PASSWORD wajib diisi bila AUTO_SEED=true", label, strings.ToUpper(label))
			}
			if len(acc.Password) < 8 || len(acc.Password) > 72 {
				return fmt.Errorf("seed %s: password harus 8-72 karakter", label)
			}
		}
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s harus bernilai true atau false", key)
	}
	return v, nil
}

func parseInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s harus berupa angka", key)
	}
	return v, nil
}

func parseInt32(key string, fallback int32) (int32, error) {
	v, err := parseInt(key, int(fallback))
	return int32(v), err
}

func parseInt64(key string, fallback int64) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s harus berupa angka", key)
	}
	return v, nil
}

func parseDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s harus berupa durasi (contoh: 24h)", key)
	}
	return v, nil
}
