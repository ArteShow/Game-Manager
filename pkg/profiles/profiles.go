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

func DeletProfile(db *sql.DB, profileID, userID int64) (bool, error) {
	result, err := db.Exec(`DELETE FROM profiles WHERE profile_id = ? AND user_id = ?`, profileID, userID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func InsertGameIntoTable(db *sql.DB, gameName string, profileID, userID int64) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM profiles WHERE profile_id = ? AND user_id = ?)",
		profileID, userID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	_, err = db.Exec("INSERT INTO games (profile_id, name) VALUES (?, ?)", profileID, gameName)
	if err != nil {
		return false, err
	}

	_, err = db.Exec("UPDATE profiles SET used_count = used_count + 1 WHERE profile_id = ?", profileID)
	if err != nil {
		return false, err
	}

	return true, nil
}
