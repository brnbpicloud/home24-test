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
	ApiAddr  string
	infoLog  *log.Logger
	errorLog *log.Logger
	Redis    RedisStore
}

type RedisStore interface {
	GetAllURLs(ctx context.Context) ([]*cache.URLTracker, error)
	GetURL(ctx context.Context, id string) (*cache.URLTracker, error)
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

	app.infoLog.Println("Starting HTTP server on port:", serverPort)

	return srv.ListenAndServe()
}

func main() {
	inforLog := log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	apiAddr := os.Getenv("API_ADDR")
	if apiAddr == "" {
		apiAddr = "http://localhost:4001"
	}
	redisClient := cache.NewRedisClient(redisAddr)
	defer redisClient.Close()

	app := &application{
		ApiAddr:  apiAddr,
		infoLog:  inforLog,
		errorLog: errorLog,
		Redis:    redisClient,
	}

	err := app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
