package setup

import (
	"crypto/rand"
	"encoding/base64"
	"log"

	"github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
)

func GenerateRandomKey(size int) (string, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(key), nil
}

func SetUpProfilesDatabase() error {
	dbPath, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		log.Fatal(err)
		return err
	}

	err2 := db.CreateDatabase(dbPath)
	if err2 != nil {
		return err2
	}

	Database, err := db.OpenDataBase(dbPath)
	if err != nil {
		log.Fatal(err)
		return err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			profile_id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			used_count INTEGER NOT NULL DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS games (
			game_id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			UNIQUE(profile_id, name)
		);`,
	}

	for _, q := range queries {
		if _, err := Database.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

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

func SetUpJWT() error {
	path, err := getconfig.GetJWTDatabasePath()
	if err != nil {
		return err
	}

	if err := db.CreateDatabase(path); err != nil {
		return err
	}

	db2, err := db.OpenDataBase(path)
	if err != nil {
		return err
	}
	defer db2.Close()

	query := `
    CREATE TABLE IF NOT EXISTS jwt (
        id      INTEGER PRIMARY KEY AUTOINCREMENT,
        jwt_key TEXT NOT NULL UNIQUE
    );`
	if _, err := db2.Exec(query); err != nil {
		return err
	}

	_, err = db2.Exec("DELETE FROM jwt")
	if err != nil {
		return err
	}

	key, err := GenerateRandomKey(32)
	if err != nil {
		return err
	}

	columns := []string{"jwt_key"}
	values := []any{key}
	if err := db.InsertValueInTable("jwt", columns, values, db2); err != nil {
		return err
	}

	return nil
}

func SetUp() error {
	err := SetUpUsersDatabase()
	if err != nil {
		log.Fatal(err)
		return err
	}

	err2 := SetUpJWT()
	if err2 != nil {
		log.Fatal(err)
		return err
	}

	err3 := SetUpProfilesDatabase()
	if err3 != nil {
		log.Fatal(err3)
		return err3
	}

	return nil
}
