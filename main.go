package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	nr "github.com/matnich89/national-rail-client/nationalrail"
	"log"
	"sync"
	cmd "trainstats-cancellations/cmd/api"
	"trainstats-cancellations/config"
	"trainstats-cancellations/db"
	http_handler "trainstats-cancellations/http/handler"
)

func main() {

	c, err := config.Load()

	if err != nil {
		log.Fatalf("could not load config %v", err)
	}

	database := db.NewCancellationDb(c.DatabaseConnection, c.DatabaseMigrationDir)

	err = database.Connect()
	if err != nil {
		log.Fatalf("Error connecting to database")
	} else {
		log.Println("connected to db :)")
	}

	defer func(database *db.CancellationDb) {
		_ = database.Close()
	}(database)

	err = database.Migrate()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	nrClient, err := nr.NewClient(
		nr.AccessTokenOpt(c.NationalRailApiKey))

	if err != nil {
		log.Fatalf("Failed to create nationalrail  client: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: c.RedisAddress,
	})

	httpHandler := http_handler.Handler{}

	app := cmd.NewApp(chi.NewMux(), &httpHandler, redisClient, nrClient, database, 2, &sync.WaitGroup{})

	app.SetupWorkers()

	app.Run()

}
