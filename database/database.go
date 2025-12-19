package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DbConfig struct {
	DbName   string
	User     string
	Password string
	URL      string
	Port     string
}

var Cfg DbConfig

func (c *DbConfig) getConfig() string {
	if c.DbName == "" {
		c.DbName = "postgres"
	}
	if c.User == "" {
		c.User = "postgres"
	}
	if c.URL == "" {
		c.URL = "localhost"
	}
	//postgresql://[user[:password]@][host][:port][/dbname][?options]
	strconnect := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		c.User,
		c.Password,
		c.URL,
		c.Port,
		c.DbName)
	fmt.Println(strconnect)
	slog.Debug(strconnect)
	return strconnect
}

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {

	pool, err := pgxpool.New(ctx, Cfg.getConfig())

	if err != nil {
		slog.Debug("error pgxpool.New")
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("error connetc db", "ping", err)
	}

	return pool, err

}
