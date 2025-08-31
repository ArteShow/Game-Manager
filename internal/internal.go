package internal

import (
	//"log"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ArteShow/Game-Manager/models"
	"github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
)

func RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	path, err := getconfig.GetUserdatabasePath()
	db, err := db.OpenDataBase(path)
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error while reading request", http.StatusInternalServerError)
		return
	}
	var UserData models.Login
	err = json.Unmarshal(body, &UserData)
	if err != nil {
		http.Error(w, "Error unmarshling the body", http.StatusInternalServerError)
		return
	}
	db.Close() //delet
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
