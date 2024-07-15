package cmd

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	nr "github.com/matnich89/national-rail-client/nationalrail"
	"log"
	"net/http"
	"sync"
	"trainstats-cancellations/consumer"
	"trainstats-cancellations/db"
	httphandler "trainstats-cancellations/http/handler"
)

type App struct {
	router      *chi.Mux
	httpHandler *httphandler.Handler
	redisClient *redis.Client
	nrClient    *nr.Client
	database    db.Database
	numWorkers  int
	consumers   []*consumer.Worker
	wg          *sync.WaitGroup
}

func NewApp(router *chi.Mux, httpHandler *httphandler.Handler,
	redisClient *redis.Client, nrClient *nr.Client, database db.Database, wg *sync.WaitGroup) *App {
	return &App{router: router, httpHandler: httpHandler, redisClient: redisClient, nrClient: nrClient, database: database, wg: wg}
}

func (a *App) SetupWorkers() {
	log.Println("setting up  workers")

	for i := 0; i < a.numWorkers; i++ {
		worker := consumer.NewConsumerWorker(i, a.nrClient, a.redisClient, a.database, a.wg)
		a.consumers = append(a.consumers, worker)
	}
}

func (a *App) Run() {
	log.Println("Starting application")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start webserver
	a.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := a.serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server error: %v", err)
		}
	}(a.wg)

	// start consumers
	for _, c := range a.consumers {
		a.wg.Add(1)
		go c.Listen(ctx)
	}

	a.wg.Wait()
	log.Println("Application shutdown complete")
}

func (a *App) routes() {
	a.router.Get("/test", func(w http.ResponseWriter, r *http.Request) {})
}
