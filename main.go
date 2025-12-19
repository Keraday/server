package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"server/handlers"
	"server/logger"
	"time"
)

func main() {
	var debug bool
	var port string
	flag.BoolVar(&debug, "debug", false, "enable debug mode(default INFO)")
	flag.StringVar(&port, "port", "8080", "on which port to run the server")
	flag.Parse()
	logger.Init(debug, os.Stdout)

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
