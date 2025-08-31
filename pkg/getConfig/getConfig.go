package getconfig

import (
	"encoding/json"
	"log"
	"os"

	"github.com/ArteShow/Game-Manager/models"
)

func GetApplicationPort() (int, error) {
	log.Println("Loading configuration from")
	configFile, err := os.Open("./configs/config.json")
	if err != nil {
		log.Println("Error opening config file:", err)
		return 0, err
	}
	defer configFile.Close()

	ports := &models.Ports{}
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(ports)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return 0, err
	}
	log.Println("Configuration loaded successfully")
	return ports.ApplicationPort, nil
}

func GetUserdatabasePath() (string, error) {
	log.Println("Loading configuration from")
	configFile, err := os.Open("./configs/config.json")
	if err != nil {
		log.Println("Error opening config file:", err)
		return "", err
	}
	defer configFile.Close()

	paths := &models.DatabasePaths{}
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(paths)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return "", err
	}
	log.Println("Configuration loaded successfully")
	return paths.UserDatabasePath, nil
}
