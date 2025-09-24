package getconfig

import (
	"encoding/json"
	"log"
	"os"

	"github.com/ArteShow/Game-Manager/models"
)

func GetTournamentPort() (int, error) {
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
	return ports.ApplicationPort, nil
}

func GetApplicationPort() (int, error) {
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
	return ports.ApplicationPort, nil
}

func GetUserdatabasePath() (string, error) {
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
	return paths.UserDatabasePath, nil
}

func GetInternalPort() (int, error) {
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
	return ports.InternalPort, nil
}

func GetJWTDatabasePath() (string, error) {
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
	return paths.JWTDatabsePath, nil
}

func GetProfilsDatabasePath() (string, error) {
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
	return paths.ProfilsDatabasePath, nil
}

func GetHubPort() (int, error) {
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
	return ports.HubPort, nil
}

func GetTasksFilePath() (string, error) {
	configFile, err := os.Open("./configs/config.json")
	if err != nil {
		log.Println("Error opening config file:", err)
		return "", err
	}
	defer configFile.Close()

	path := &models.TasksPath{}
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(path)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return "", err
	}
	return path.TasksPath, nil
}

func GetStaticFolderPath() (string, error) {
	configFile, err := os.Open("./configs/config.json")
	if err != nil {
		log.Println("Error opening config file:", err)
		return "", err
	}
	defer configFile.Close()

	paths := &models.StaticFolderPath{}
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(paths)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return "", err
	}
	return paths.Static, nil
}
