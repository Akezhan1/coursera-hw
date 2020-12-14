package main

import (
	"database/sql"
	"net/http"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type dbExplorer struct {
	DB     *sql.DB
	Tables map[string]string
}

func NewDbExplorer(db *sql.DB) (*dbExplorer, error) {
	return &dbExplorer{DB: nil, Tables: nil}, nil
}

func (d *dbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
