package db

import (
	"database/sql"

	"integration-education-db/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

func NewConnection(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DatabaseDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Database.MaxConnections)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.Database.ConnectionLifetime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
