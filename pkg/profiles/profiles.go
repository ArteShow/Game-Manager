package profiles

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ArteShow/Game-Manager/models"
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
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM profiles WHERE profile_id = ? AND user_id = ?",
		profileID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, fmt.Errorf("profile does not belong to user")
	}

	_, err = db.Exec("INSERT INTO games (profile_id, name) VALUES (?, ?)", profileID, gameName)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return false, fmt.Errorf("game '%s' already exists for this profile", gameName)
		}
		return false, err
	}

	return true, nil
}

func GetAllUsersProfiles(db *sql.DB, userID int64) ([]models.ProfileData, error) {
	rows, err := db.Query(
		"SELECT profile_id, user_id, name, description, used_count FROM profiles WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []models.ProfileData

	for rows.Next() {
		var p models.ProfileData
		err := rows.Scan(&p.ProfileID, &p.UserID, &p.Name, &p.Description, &p.UsedCount)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func DeletGame(db *sql.DB, profileID, userID, gameID int64) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM profiles WHERE profile_id = ? AND user_id = ?",
		profileID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	res, err := db.Exec("DELETE FROM games WHERE game_id = ? AND profile_id = ?", gameID, profileID)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, nil
	}

	return true, nil
}

func GetAllGamesFromAProfile(db *sql.DB, profileID, userID int64) ([]models.Game, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM profiles WHERE profile_id = ? AND user_id = ?",
		profileID, userID,
	).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("profile does not belong to user")
	}

	rows, err := db.Query("SELECT game_id, profile_id, name FROM games WHERE profile_id = ?", profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var game models.Game
		if err := rows.Scan(&game.GameID, &game.ProfileID, &game.Name); err != nil {
			return nil, err
		}
		game.UserID = userID
		games = append(games, game)
	}

	return games, nil
}

func GetGameByGameIDProfileIDUserID(db *sql.DB, profileID, userID, gameID int64) (models.Game, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM profiles WHERE profile_id = ? AND user_id = ?",
		profileID, userID,
	).Scan(&count)
	if err != nil {
		return models.Game{}, err
	}
	if count == 0 {
		return models.Game{}, fmt.Errorf("profile does not belong to user")
	}

	var game models.Game
	err = db.QueryRow(
		"SELECT game_id, profile_id, name FROM games WHERE game_id = ? AND profile_id = ?",
		gameID, profileID,
	).Scan(&game.GameID, &game.ProfileID, &game.Name)
	if err != nil {
		return models.Game{}, err
	}

	game.UserID = userID
	return game, nil
}
