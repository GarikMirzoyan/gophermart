package main

import (
	"log"

	"github.com/GarikMirzoyan/gophermart/internal/app"
)

func main() {
	appInstance, err := app.New()
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	err = appInstance.Run()
	if err != nil {
		log.Fatalf("error running app: %v", err)
	}

	// Например, appInstance.Run() для старта сервера
	log.Printf("App initialized with config: %+v", appInstance.Config)
}
