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
	migrateDir := "/Users/mathewnicholls/repo/trainstats-cancellations/db/migrations"

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

	t.Run("UpdateAndGetValue", func(t *testing.T) {
		date := time.Now().Format("2006-01-02")
		value := 42

		err := testDb.UpdateValue(date, value)
		assert.NoError(t, err)

		retrievedValue, err := testDb.GetValue(date)
		assert.NoError(t, err)
		assert.Equal(t, value, retrievedValue)

		nonExistentDate := "2000-01-01"
		retrievedValue, err = testDb.GetValue(nonExistentDate)
		assert.NoError(t, err)
		assert.Equal(t, 0, retrievedValue)
	})

	t.Run("UpdateValue_Conflict", func(t *testing.T) {
		date := time.Now().Format("2006-01-02")
		firstValue := 1
		secondValue := 2

		err := testDb.UpdateValue(date, firstValue)
		assert.NoError(t, err)

		err = testDb.UpdateValue(date, secondValue)
		assert.NoError(t, err)

		retrievedValue, err := testDb.GetValue(date)
		assert.NoError(t, err)
		assert.Equal(t, secondValue, retrievedValue)
	})

}
