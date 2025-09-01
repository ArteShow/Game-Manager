package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ArteShow/Game-Manager/models"
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
		log.Fatal(err)
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
		log.Fatal(err)
		return
	}

	db, err := DB.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	profileID, err := profiles.CreateProfileInDB(db, profileData.UserID, profileData.Name, profileData.Description)
	if err != nil {
		http.Error(w, "Failed to insert or get the profileID", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	log.Println("New profile created with ID:", profileID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"profile_id": %d}`, profileID)))
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

	return http.ListenAndServe(portStr, nil)
}
