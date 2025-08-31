package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDataBase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
		return db, err
	}
	return db, nil
}

func CreateDatabase(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}
	defer db.Close()
	return nil
}

func GetValueFromTableByID(db *sql.DB, table string, column string, id int) (any, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", column, table)

	var value any
	err := db.QueryRow(query, id).Scan(&value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func InsertValueInTable(table string, columns []string, values []any, db *sql.DB) error {
	placeholders := ""
	for i := range values {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		joinColumns(columns),
		placeholders,
	)

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(values...)
	return err
}

func joinColumns(cols []string) string {
	result := ""
	for i, col := range cols {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

func DeleteValueByID(db *sql.DB, table string, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)

	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}
