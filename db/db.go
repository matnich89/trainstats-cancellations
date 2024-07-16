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
	"time"
)

const (
	insertCancellationSQL = `
       INSERT INTO cancellations (train_id, operator, cancellation_date, cancellation_reason)
       VALUES ($1, $2, $3, $4)
    `
	getCancellationCountSQL = `SELECT COUNT(*) FROM cancellations WHERE DATE(cancellation_date) = $1`
)

type Database interface {
	Connect() error
	Close() error
	Migrate() error
	InsertCancellation(trainID, operator string, cancellationDate time.Time, cancellationReason string) error
	GetCancellationCountForDate(date time.Time) (int, error)
}

type CancellationDb struct {
	db         *sql.DB
	migrateDir string
	ConnStr    string
}

func NewCancellationDb(connStr, migrateDir string) *CancellationDb {
	return &CancellationDb{
		ConnStr:    connStr,
		migrateDir: migrateDir,
	}
}

func (c *CancellationDb) Connect() error {
	db, err := sql.Open("postgres", c.ConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	c.db = db
	return nil
}

func (c *CancellationDb) Close() error {
	return c.db.Close()
}

func (c *CancellationDb) InsertCancellation(trainID, operator string, cancellationDate time.Time, cancellationReason string) error {
	res, err := c.db.Exec(insertCancellationSQL, trainID, operator, cancellationDate, cancellationReason)
	log.Println(res)
	if err != nil {
		return fmt.Errorf("failed to insert cancellation: %w", err)
	}
	return nil
}

func (c *CancellationDb) GetCancellationCountForDate(date time.Time) (int, error) {
	var count int
	err := c.db.QueryRow(getCancellationCountSQL, date.Format("2006-01-02")).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get cancellation count: %w", err)
	}
	return count, nil
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
