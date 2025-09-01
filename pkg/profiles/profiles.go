package profiles

import (
	"database/sql"
	"fmt"
)

func CreateProfileInDB(db *sql.DB, userID int64, name string, description string) (int64, error) {
	query := `INSERT INTO profiles (user_id, name, description) VALUES (?, ?, ?)`
	result, err := db.Exec(query, userID, name, description)
	if err != nil {
		return 0, fmt.Errorf("failed to insert profile: %w", err)
	}

	profileID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get profile ID: %w", err)
	}

	return profileID, nil
}
