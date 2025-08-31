package setup

import (
	"log"

	"github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
)

func SetUpUsersDatabase() error {
	path, err := getconfig.GetUserdatabasePath()
	if err != nil {
		log.Fatal(err)
		return err
	}

	err3 := db.CreateDatabase(path)
	if err3 != nil {
		log.Fatal(err)
		return err
	}

	db, err := db.OpenDataBase(path)
	if err != nil {
		log.Fatal(err)
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS users (
		user_id   INTEGER PRIMARY KEY AUTOINCREMENT,
		username  TEXT NOT NULL UNIQUE,
		password  TEXT NOT NULL
	);`

	_, err2 := db.Exec(query)
	if err2 != nil {
		log.Fatal(err2)
		return err2
	}
	return nil
}

func SetUp() error {
	err := SetUpUsersDatabase()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
