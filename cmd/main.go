package main

import (
	"log"

	"github.com/ArteShow/Game-Manager/application"
	"github.com/ArteShow/Game-Manager/internal"
	"github.com/ArteShow/Game-Manager/pkg/hub"
	"github.com/ArteShow/Game-Manager/pkg/setup"
	"github.com/ArteShow/Game-Manager/pkg/tournament/live"
)

func main() {

	err := setup.SetUp()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	go func() {
		log.Println("Loading Application Server")
		err := application.StartApplicationServer()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
	}()

	go func() {
		log.Println("Loading Internla Server")
		err := internal.StartInternalServer()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
	}()

	go func() {
		log.Println("Loading Hub Server")
		err := hub.StartHub()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
	}()

	go func ()  {
		log.Println("Loading Tournament Server")
		err := live.StartLive()
		if err != nil{
			log.Fatal(err)
			panic(err)
		}
	}()

	select {}
}
