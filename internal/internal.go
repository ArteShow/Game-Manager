package internal

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ArteShow/Game-Manager/models"
	DB "github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
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
		http.Error(w, "Error unmarshling the body", http.StatusInternalServerError)
		return
	}
	log.Printf("Login attempt: username=%s\n", UserData.Username)

	yesORno, err := registartion.CheckUserData(UserData)
	if err != nil {
		http.Error(w, "Internal check failed", http.StatusInternalServerError)
		return
	}

	if yesORno {
		key, err := registartion.GenerateJWT(UserData.UserID)
		if err != nil {
			http.Error(w, "Failed to generate a new key", http.StatusInternalServerError)
			return
		}

		bytes := []byte(key)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	} else {
		http.Error(w, "Wrong Username or password", http.StatusUnauthorized)
	}
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

	return http.ListenAndServe(portStr, nil)
}
