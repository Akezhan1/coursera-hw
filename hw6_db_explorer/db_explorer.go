package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type response map[string]interface{}

//DBExplorer ...
type DBExplorer struct {
	DB     *sql.DB
	Tables []*Table
}

//Table ...
type Table struct {
	Name    string
	Columns []*Column
}

//Column ...
type Column struct {
	Name       sql.NullString
	Type       sql.NullString
	PrimaryKey sql.NullBool
	AutoInc    sql.NullBool
	Null       sql.NullBool
	Default    interface{}
}

func NewDbExplorer(db *sql.DB) (*DBExplorer, error) {
	tables := make([]*Table, 0)
	tablesName := make([]string, 0)
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		rows.Scan(&table)
		tablesName = append(tablesName, table)
	}

	for _, table := range tablesName {
		rows, err := db.Query("SHOW FULL COLUMNS FROM " + table)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		columns := make([]*Column, 0)

		rowsColumns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		count := len(rowsColumns)
		values := make([]interface{}, count)
		scanArgs := make([]interface{}, count)
		for i := range values {
			scanArgs[i] = &values[i]
		}

		for rows.Next() {
			err := rows.Scan(scanArgs...)
			if err != nil {
				return nil, err
			}
			column := &Column{}
			for i, col := range rowsColumns {
				v := values[i]
				b, _ := v.([]byte)
				switch col {
				case "Field":
					column.Name.Scan(b)
				case "Type":
					column.Type.Scan(b)
				case "Null":
					column.Null.Scan(string(b) == "YES")
				case "Key":
					column.PrimaryKey.Scan(string(b) == "PRI")
				case "Extra":
					column.AutoInc.Scan(string(b) == "auto_increment")
				case "Default":
					column.Default = v
				}
			}
			columns = append(columns, column)
		}
		tables = append(tables, &Table{Name: table, Columns: columns})
	}
	return &DBExplorer{DB: db, Tables: tables}, nil
}

func (dbExp *DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path == "/" {
		if r.Method == http.MethodGet {
			dbExp.handlerGetAllTables(w, r)
		} else {
			response := response{
				"error": "bad method",
			}
			writeResponse(w, response, http.StatusBadRequest)
		}
	}
}

func (dbExp *DBExplorer) handlerGetAllTables(w http.ResponseWriter, r *http.Request) {
	tables := make([]string, 0, len(dbExp.Tables))

	for _, t := range dbExp.Tables {
		tables = append(tables, t.Name)
	}

	response := response{
		"response": map[string][]string{
			"tables": tables,
		},
	}
	writeResponse(w, response, http.StatusOK)
}

func writeResponse(w http.ResponseWriter, res response, status int) {
	data, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(data)
}
