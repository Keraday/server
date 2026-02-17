package main

import (
	"context"
	"net/http"
	"os"
	"server/config"
	"server/database"
	"server/handlers"
	"server/logger"
	"time"
)

func main() {
	cfg := config.Config()

	log := logger.New(cfg.LogLvl, cfg.ENV == "prod", os.Stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dbPool, err := database.NewPool(ctx, cfg.DBURL, log)
	if err != nil {
		log.Error("connection to the database could not be established", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	app := handlers.SubscriptionHandler{
		DB:  dbPool,
		Log: log,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /subscriptions", app.Create)
	mux.HandleFunc("GET /subscriptions/{id}", app.GetByID)
	mux.HandleFunc("DELETE /subscriptions/{id}", app.Delete)
	mux.HandleFunc("PATCH /subscriptions/{id}", app.Update)
	mux.HandleFunc("GET /subscriptions/total", app.GetTotal)

	serv := &http.Server{
		Addr:         cfg.ServAddr,
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 6 * time.Second,
		IdleTimeout:  20 * time.Second,
	}
	// TODO delete pass form dburl
	log.Debug("Config Set: ", "ENV:", cfg.ENV, "DB_URL:", cfg.DBURL, "Log_Lvl:", cfg.LogLvl, "SERV_ADDR:", cfg.ServAddr)
	log.Info("Server UP")
	if err := serv.ListenAndServe(); err != nil {
		log.Error("error start server ", "error", err)
		os.Exit(1)
	}
}
