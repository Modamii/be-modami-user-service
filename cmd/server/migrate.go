package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/modami/user-service/config"
	"github.com/modami/user-service/migrations"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

func ensureDatabase(ctx context.Context, cfg config.PostgresConfig) error {
	// Connect to the default "postgres" database to create the target DB.
	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		cfg.UserWriter, cfg.PasswordWriter, cfg.Host, cfg.Port)

	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer db.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	if err := db.QueryRowContext(ctx, query, cfg.Database).Scan(&exists); err != nil {
		return fmt.Errorf("check database existence: %w", err)
	}

	if exists {
		return nil
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.Database)); err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	logger.Info(ctx, "database created", logging.String("database", cfg.Database))
	return nil
}

func runMigrations(ctx context.Context, dsn string) error {
	d, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	version, _, _ := m.Version()
	logger.Info(ctx, "migrations applied", logging.Int("version", int(version)))
	return nil
}
