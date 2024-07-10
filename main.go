package main

import (
	"log"
	"trainstats-cancellations/db"
)

func main() {
	connStr := "postgres://postgres:postgres@localhost/cancellations?sslmode=disable"
	migrateDir := "./db/migrations"
	database := db.NewCancellationDb(connStr, migrateDir)

	err := database.Connect()
	if err != nil {
		log.Println("Error connecting to database")
	} else {
		log.Println("connected to db :)")
	}
	defer database.Close()

	err = database.Migrate()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

}
