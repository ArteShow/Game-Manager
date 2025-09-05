package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ArteShow/Game-Manager/models"
	"github.com/ArteShow/Game-Manager/pkg/db"
	DB "github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/ArteShow/Game-Manager/pkg/profiles"
	"github.com/ArteShow/Game-Manager/pkg/registartion"
)

func RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Register new user in Internal")

	path, err := getconfig.GetUserdatabasePath()
	if err != nil {
		http.Error(w, "Error while getting the path for the databse", http.StatusInternalServerError)
		return
	}

	db, err := DB.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var UserData models.Login
	err = json.Unmarshal(body, &UserData)
	if err != nil {
		http.Error(w, "Error unmarshling the body", http.StatusInternalServerError)
		return
	}
	log.Printf("User data received: username=%s\n", UserData.Username)

	columns := []string{"username", "password"}
	values := []any{UserData.Username, UserData.Passwword}

	err = DB.InsertValueInTable("users", columns, values, db)
	if err != nil {
		http.Error(w, "Error while inserting into users new user", http.StatusInternalServerError)
		return
	}

	log.Println("User successfully inserted into DB")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func LoginAUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request received")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var UserData models.Login
	err = json.Unmarshal(body, &UserData)
	if err != nil {
		http.Error(w, "Error unmarshaling the body", http.StatusInternalServerError)
		return
	}
	log.Printf("Login attempt: username=%s\n", UserData.Username)

	userID, err := registartion.GetUserIDByCredentials(UserData.Username, UserData.Passwword)
	if err != nil {
		http.Error(w, "Wrong Username or password", http.StatusUnauthorized)
		return
	}
	key, err := registartion.GenerateJWT(userID)
	if err != nil {
		http.Error(w, "Failed to generate a new key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(key))
}

func CreateProflie(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating new Profile")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var profileData models.ProfileData
	err = json.Unmarshal(body, &profileData)
	if err != nil {
		http.Error(w, "Error unmarshaling the body", http.StatusInternalServerError)
		return
	}

	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "Failed to get the path", http.StatusInternalServerError)
		return
	}

	db, err := DB.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}

	profileID, err := profiles.CreateProfileInDB(db, profileData.UserID, profileData.Name, profileData.Description)
	if err != nil {
		http.Error(w, "Failed to insert or get the profileID", http.StatusInternalServerError)
		return
	}

	log.Println("New profile created with ID:", profileID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"profile_id": %d}`, profileID)))
}

func DeletProfile(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var profileData models.ProfileData
	err = json.Unmarshal(body, &profileData)
	if err != nil {
		http.Error(w, "Error unmarshaling the body", http.StatusInternalServerError)
		return
	}

	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "error while getting the path for the database", http.StatusInternalServerError)
		return
	}

	db, err := db.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}

	ok, err := profiles.DeletProfile(db, profileData.ProfileID, profileData.UserID)
	if err != nil {
		http.Error(w, "Failed to delet profile", http.StatusInternalServerError)
		return
	}

	if ok {
		log.Println("Profile deletes sucsessfuly")
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Failed to delet")
	}
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var Game models.Game
	err = json.Unmarshal(body, &Game)
	if err != nil {
		http.Error(w, "Error unmarshaling the body", http.StatusInternalServerError)
		return
	}

	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "error while getting the path for the database", http.StatusInternalServerError)
		return
	}

	db, err := db.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}

	ok, err := profiles.InsertGameIntoTable(db, Game.Name, Game.ProfileID, Game.UserID)
	if err != nil {
		http.Error(w, "Failed to insert value in the database", http.StatusInternalServerError)
		return
	}

	if !ok {
		log.Println("Wrong UserID or ProfileID")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func GetAllUsersProflies(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var Profile models.ProfileData
	err = json.Unmarshal(body, &Profile)
	if err != nil {
		http.Error(w, "Error unmarshaling the body", http.StatusInternalServerError)
		return
	}

	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "Error while getting the path for the database", http.StatusInternalServerError)
		return
	}

	db, err := DB.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	Profiles, err := profiles.GetAllUsersProfiles(db, Profile.UserID)
	if err != nil {
		http.Error(w, "Failed to get Profiles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(Profiles)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func DeletGame(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var Game models.Game
	err = json.Unmarshal(body, &Game)
	if err != nil {
		http.Error(w, "Error unmarshaling the body", http.StatusInternalServerError)
		return
	}

	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "error while getting the path for the database", http.StatusInternalServerError)
		return
	}

	db, err := db.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}

	ok, err := profiles.DeletGame(db, Game.ProfileID, Game.UserID, Game.GameID)
	if err != nil {
		http.Error(w, "Failed to delet the game", http.StatusInternalServerError)
		return
	}

	if !ok {
		log.Println("Something went wrong pls try again")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func GetUsersGames(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var gameReq models.Game
	err = json.Unmarshal(body, &gameReq)
	if err != nil {
		http.Error(w, "Error unmarshaling request body", http.StatusInternalServerError)
		return
	}
	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "Failed to get database path", http.StatusInternalServerError)
		return
	}

	db, err := db.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	game, err := profiles.GetAllGamesFromAProfile(db, gameReq.ProfileID, gameReq.UserID)
	if err != nil {
		http.Error(w, "Failed to fetch games", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	respBytes, err := json.Marshal(game)
	if err != nil {
		http.Error(w, "Failed to marshal games to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func GetUsersGameByIDAndProfileID(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var gameReq models.Game
	err = json.Unmarshal(body, &gameReq)
	if err != nil {
		http.Error(w, "Error unmarshaling request body", http.StatusInternalServerError)
		return
	}
	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "Failed to get database path", http.StatusInternalServerError)
		return
	}

	db, err := db.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	games, err := profiles.GetGameByGameIDProfileIDUserID(db, gameReq.ProfileID, gameReq.UserID, gameReq.GameID)
	if err != nil {
		http.Error(w, "Failed to fetch games", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	respBytes, err := json.Marshal(games)
	if err != nil {
		http.Error(w, "Failed to marshal games to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func ChooseProfile(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var profileData models.ProfileData
	err = json.Unmarshal(body, &profileData)
	if err != nil {
		http.Error(w, "Error unmarshaling request body", http.StatusInternalServerError)
		return
	}
	path, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		http.Error(w, "Failed to get database path", http.StatusInternalServerError)
		return
	}

	db, err := db.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	profile, err := profiles.GetProfileByID(db, profileData.UserID, profileData.ProfileID)
	if err != nil {
		http.Error(w, "Failed to fetch games", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	respBytes, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, "Failed to marshal games to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func StartInternalServer() error {
	log.Println("Starting Internal Server")
	port, err := getconfig.GetInternalPort()
	if err != nil {
		log.Fatal(err)
	}
	portStr := ":" + strconv.Itoa(port)

	http.HandleFunc("/internal/register", RegisterNewUser)
	http.HandleFunc("/internal/login", LoginAUser)

	http.HandleFunc("/internal/createProfile", CreateProflie)
	http.HandleFunc("/internal/deletProfile", DeletProfile)
	http.HandleFunc("/internal/createGame", CreateGame)
	http.HandleFunc("/internal/getUsersProfiles", GetAllUsersProflies)
	http.HandleFunc("/internal/deletGame", DeletGame)
	http.HandleFunc("/internal/getGames", GetUsersGames)
	http.HandleFunc("/internal/getGameById", GetUsersGameByIDAndProfileID)
	http.HandleFunc("/internal/chooseProfile", ChooseProfile)

	return http.ListenAndServe(portStr, nil)
}
