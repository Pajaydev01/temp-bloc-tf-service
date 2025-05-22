package init

import (
	"log"
	"os"
	"time"

	"github.com/bloc-transfer-service/pkg/transfers/api"
	QueueHandler "github.com/bloc-transfer-service/utils/AsyncQueue"
	"github.com/bloc-transfer-service/utils/easypay"
	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
)

// Init function for Transfers
func InitTransfer(router *mux.Router) {
	api.Router(router)
	go func() {
		server := asynq.NewServer(
			asynq.RedisClientOpt{
				Addr:        os.Getenv("REDIS_ADDR"),
				Password:    os.Getenv("REDIS_PASSWORD"),
				DialTimeout: time.Second * 20,
				DB:          0,
				Username:    os.Getenv("REDIS_USER")},
			asynq.Config{
				Concurrency: 4, // Number of workers
			},
		)
		route := asynq.NewServeMux()
		route.HandleFunc(QueueHandler.ProcessTransfer, easypay.HandleTsqQueue)
		if err := server.Run(route); err != nil {
			log.Fatalf("Could not start server: %v", err)
		}
	}()
}
