package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"server/database"
	"server/handlers"
	"server/logger"
	"time"
)

func main() {
	var debug bool
	var port string
	flag.BoolVar(&debug, "debug", true, "enable debug mode(default INFO)")
	flag.StringVar(&port, "port", "8080", "on which port to run the server")
	flag.StringVar(&database.Cfg.DbName, "postgres", "", "login postgres")
	flag.StringVar(&database.Cfg.Password, "password", "", "password postgres ")
	flag.StringVar(&database.Cfg.URL, "url", "localhost", "uri postgres")
	flag.StringVar(&database.Cfg.Port, "db-port", "5432", "on which port to run the DB")

	flag.Parse()

	logger.Init(debug, os.Stdout)

	ctx := context.TODO()

	db, err := database.NewPool(ctx)
	if err != nil {
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.Handler1)

	serv := &http.Server{
		Addr:         "localhost:" + port,
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 6 * time.Second,
		IdleTimeout:  20 * time.Second,
	}
	slog.Info("Server UP")
	if err := serv.ListenAndServe(); err != nil {
		slog.Error("error start server ", slog.Any("error", err))
		os.Exit(1)
	}
}
