package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"urltracker/internal"
	"urltracker/internal/cache"
)

type RedisStore interface {
	DequeueURL(ctx context.Context) (*cache.URLTracker, error)
	UpdateURL(ctx context.Context, t *cache.URLTracker) error
}

type CrawlerFunc func(t *cache.URLTracker) (string, error)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	interval := 5 * time.Second

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	r := cache.NewRedisClient(redisAddr)
	logger := log.New(os.Stdout, "[worker] ", log.LstdFlags)

	logger.Println("worker starting, redis:", redisAddr)

	for {
		select {
		case <-ctx.Done():
			logger.Println("shutting down")
			return
		default:
			for {
				processed, err := processNext(ctx, r, CrawlURL, logger)
				if err != nil {
					logger.Println("dequeue error:", err)
					break
				}
				if !processed {
					// queue empty
					break
				}
			}

			time.Sleep(interval)
		}
	}
}

func processNext(ctx context.Context, store RedisStore, crawl CrawlerFunc, logger *log.Logger) (bool, error) {
	tracker, err := store.DequeueURL(ctx)
	if err != nil {
		return false, err
	}
	if tracker == nil {
		return false, nil
	}

	tracker.Status = internal.StatusProcessing
	tracker.UpdatedAt = time.Now()
	if err := store.UpdateURL(ctx, tracker); err != nil {
		logger.Println("update processing status:", err)
		return true, nil
	}

	result, cErr := crawl(tracker)

	if cErr != nil {
		tracker.Status = internal.StatusFailed
		tracker.Error = cErr.Error()
	} else {
		tracker.Status = internal.StatusCompleted
		tracker.Result = result
	}
	tracker.UpdatedAt = time.Now()

	if err := store.UpdateURL(ctx, tracker); err != nil {
		logger.Println("update final status:", err)
	} else {
		logger.Println("processed", tracker.ID, tracker.Status)
	}

	return true, nil
}
