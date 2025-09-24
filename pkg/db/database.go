package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDataBase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
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

func GetValueFromTableByID(db *sql.DB, table string, column string, id int64, idOF string) (any, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", column, table, idOF)

	var value any
	err := db.QueryRow(query, id).Scan(&value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func InsertValueInTable(table string, columns []string, values []any, db *sql.DB) error {
	if len(columns) != len(values) {
		return fmt.Errorf("columns and values length mismatch")
	}

	colNames := strings.Join(columns, ", ")

	placeholders := strings.Repeat("?, ", len(values))
	placeholders = strings.TrimSuffix(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, colNames, placeholders)

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(values...)
	return err
}

func GetCurrentJWTKey(db *sql.DB) (string, error) {
	query := "SELECT jwt_key FROM jwt LIMIT 1"

	var key string
	err := db.QueryRow(query).Scan(&key)
	if err != nil {
		return "", err
	}

	return key, nil
}

func DeleteValueByID(db *sql.DB, table string, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)

	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}
