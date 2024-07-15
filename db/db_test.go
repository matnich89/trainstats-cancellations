package db_test

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"
	"trainstats-cancellations/db"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDb     *db.CancellationDb
	dbName     = "testdb"
	dbUser     = "testuser"
	dbPassword = "testpass"
)

func setupTestDatabase(t *testing.T) (func(), error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
		},
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %v", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), dbUser, dbPassword, dbName)
	migrateDir := "./migrations"

	testDb = db.NewCancellationDb(connStr, migrateDir)
	err = testDb.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = testDb.Migrate()
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		if err := testDb.Close(); err != nil {
			t.Logf("Failed to close database connection: %v", err)
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return cleanup, nil
}

func TestCancellationDb(t *testing.T) {
	cleanup, err := setupTestDatabase(t)
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer cleanup()

	t.Run("InsertCancellationAndGetCount", func(t *testing.T) {
		date := time.Now().UTC().Truncate(24 * time.Hour)

		for i := 0; i < 3; i++ {
			trainID := fmt.Sprintf("TRAIN-%d", i+1)
			err := testDb.InsertCancellation(trainID, "TestOperator", date, "Test Reason")
			assert.NoError(t, err)
		}

		count, err := testDb.GetCancellationCountForDate(date)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)

		differentDate := date.AddDate(0, 0, -1)
		count, err = testDb.GetCancellationCountForDate(differentDate)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("InsertCancellationsForMultipleDates", func(t *testing.T) {
		baseDate := time.Now().UTC().AddDate(0, -1, 0).Truncate(24 * time.Hour)

		for i := 0; i < 3; i++ {
			date := baseDate.AddDate(0, 0, i)
			trainID := fmt.Sprintf("TRAIN-%d", i+1)
			err := testDb.InsertCancellation(trainID, "TestOperator", date, "Test Reason")
			assert.NoError(t, err)
		}

		for i := 0; i < 3; i++ {
			date := baseDate.AddDate(0, 0, i)
			count, err := testDb.GetCancellationCountForDate(date)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
		}

		noCanDate := baseDate.AddDate(0, 0, -1)
		count, err := testDb.GetCancellationCountForDate(noCanDate)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
