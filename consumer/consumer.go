package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	nr "github.com/matnich89/national-rail-client/nationalrail"
	"log"
	"sync"
	"time"
	"trainstats-cancellations/db"
	"trainstats-cancellations/model"
)

type Worker struct {
	id          int
	nrClient    *nr.Client
	redisClient *redis.Client
	database    db.Database
	wg          *sync.WaitGroup
}

func NewConsumerWorker(id int, nrClient *nr.Client, redisClient *redis.Client, database db.Database, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:          id,
		nrClient:    nrClient,
		redisClient: redisClient,
		database:    database,
		wg:          wg,
	}
}

func (w *Worker) Listen(ctx context.Context) {
	defer w.wg.Done()
	for {
		msg, err := w.redisClient.BRPop(ctx, 0*time.Second, "departures-queue").Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf("Worker %d: Context canceled, exiting", w.id)
				return
			}
			log.Printf("Worker %d: Error popping from queue: %v", w.id, err)
			continue
		}

		if len(msg) != 2 {
			log.Printf("Worker %d: Unexpected result format", w.id)
			continue
		}

		var departureId model.DepartingTrainId
		err = json.Unmarshal([]byte(msg[1]), &departureId)

		if err != nil {
			log.Printf("Worker %d: Error unmarshalling departure: %v", w.id, err)
			continue
		}

		serviceDetails, err := w.nrClient.GetServiceDetails(departureId.ID)

		if err != nil {
			log.Printf("Worker %d: Error getting service details: %v", w.id, err)
			continue
		}

		if serviceDetails.IsCancelled {
			log.Println("service cancelled persisting...")
			var reason string = "NO REASON GIVEN"
			if serviceDetails.CancelReason != nil {
				reason = *serviceDetails.CancelReason
			}
			err := w.database.InsertCancellation(departureId.ID, serviceDetails.Operator.Name, time.Now().Truncate(24*time.Hour), reason)
			if err != nil {
				log.Printf("Worker %d: Error migrating database: %v", w.id, err)
			}
		} else {
			log.Println("train not cancelled")
		}
	}

}
