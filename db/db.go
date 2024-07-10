package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
)

const (
	upsertValueSQL = `
		INSERT INTO cancellations (date, value)
		VALUES ($1, $2)
		ON CONFLICT (date) DO UPDATE SET value = $2
	`
	getValueSQL = `SELECT value FROM cancellations WHERE date = $1`
)

type Database interface {
	Connect() error
	Close() error
	Migrate() error
}

type CancellationDb struct {
	db         *sql.DB
	migrateDir string
	connStr    string
}

func NewCancellationDb(connStr, migrateDir string) *CancellationDb {
	return &CancellationDb{
		connStr:    connStr,
		migrateDir: migrateDir,
	}
}

func (c *CancellationDb) Connect() error {
	db, err := sql.Open("postgres", c.connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	c.db = db
	return nil
}

func (c *CancellationDb) Close() error {
	return c.db.Close()
}

func (c *CancellationDb) UpdateValue(date string, value int) error {
	_, err := c.db.Exec(upsertValueSQL, date, value)
	if err != nil {
		return fmt.Errorf("failed to update cancellation value: %w", err)
	}
	return nil
}

func (c *CancellationDb) GetValue(date string) (int, error) {
	var value int
	err := c.db.QueryRow(getValueSQL, date).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get cancellation value: %w", err)
	}
	return value, nil
}

func (c *CancellationDb) Migrate() error {
	driver, err := postgres.WithInstance(c.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", c.migrateDir),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %v", err)
	} else if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No migration changes, done")
	} else {
		log.Println("migration changes applied, done")
	}
	return nil
}
