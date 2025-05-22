package QueueHandler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

const (
	ProcessTransfer = "processTransfer"
)

// Task payload structure
type TaskPayload struct {
	Name  string
	Items map[string]interface{} `json:"items"`
}

// Function to enqueue tasks
func EnqueueTask(payload TaskPayload, delay time.Duration) error {
	log.Println("............enqueuing Task now.......", payload.Name)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %v", err)
	}
	task := asynq.NewTask(ProcessTransfer, payloadBytes)

	// Options for delay
	opts := []asynq.Option{
		asynq.ProcessIn(delay),
	}
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:        os.Getenv("REDIS_ADDR"),
		Password:    os.Getenv("REDIS_PASSWORD"),
		DialTimeout: time.Second * 20,
		DB:          0,
		Username:    os.Getenv("REDIS_USER")})
	defer client.Close()
	_, err = client.Enqueue(task, opts...)
	return err
}
