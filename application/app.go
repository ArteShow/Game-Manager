package application

import (
	"log"
	"net/http"
	"strconv"

	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	//"log"
	//"io"
)

func RegistereNewUser(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8081/internal/reg", "application/json", r.Body)
	if err != nil {
		http.Error(w, "Error while creating new user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
}

func StartApplicationServer() error {
	log.Println("Starting Application Server")
	port, err := getconfig.GetApplicationPort()
	if err != nil {
		log.Fatal(err)
	}
	portStr := ":" + strconv.Itoa(port)

	http.HandleFunc("/reg", RegistereNewUser)

	return http.ListenAndServe(portStr, nil)
}
