package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"urltracker/internal/cache"
)

var serverPort, _ = strconv.Atoi(os.Getenv("SERVER_PORT"))

type application struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	Redis    RedisStore
}

type RedisStore interface {
	StoreURL(ctx context.Context, tracker *cache.URLTracker) error
	GetURL(ctx context.Context, id string) (*cache.URLTracker, error)
	GetAllURLs(ctx context.Context) ([]*cache.URLTracker, error)
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", serverPort),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Println("Starting backend server on port", serverPort)

	return srv.ListenAndServe()
}

func main() {
	inforLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	redisClient := cache.NewRedisClient(redisAddr)
	defer redisClient.Close()

	app := &application{
		infoLog:  inforLog,
		errorLog: errorLog,
		Redis:    redisClient,
	}

	err := app.serve()
	if err != nil {
		log.Fatal(err)
	}
}
