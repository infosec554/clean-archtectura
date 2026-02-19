package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/infosec554/clean-archtectura/config"
)

type Store struct {
	DB *sql.DB
}

func New(ctx context.Context, cfg config.Config) (*Store, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open error: %w", err)
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping error: %w", err)
	}

	cwd, _ := os.Getwd()
	mPath := filepath.Join(cwd, "migrations", "postgres")

	m, err := migrate.New("file://"+mPath, dsn)
	if err != nil {
		return nil, fmt.Errorf("migration init error: %w", err)
	}

	if err = m.Up(); err != nil {
		if strings.Contains(err.Error(), "no change") {
			fmt.Println("⚠️  No new migrations to apply")
		} else {
			return nil, fmt.Errorf("migration up error: %w", err)
		}
	} else {
		fmt.Println("✅ Migrations applied successfully")
	}

	fmt.Println("✅ PostgreSQL connected successfully")
	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}
