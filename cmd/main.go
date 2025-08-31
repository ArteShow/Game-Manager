package main

import (
	"log"

	"github.com/ArteShow/Game-Manager/application"
)

func main() {

	go func() {
		log.Println("Loading Application Server")
		err := application.StartApplicationServer()
		if err != nil {
			log.Fatal(err)
		}
	}()
}
