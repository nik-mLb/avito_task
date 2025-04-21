//	@title			pvzAPI
//	@version		1.0
//	@description	API for pvz
//	@host			localhost:8080
package main

import (
	"log"
	"github.com/nik-mLb/avito_task/config"
	"github.com/nik-mLb/avito_task/internal/app"
	_ "github.com/lib/pq"
	_ "github.com/nik-mLb/avito_task/docs"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	application, err := app.NewApp(conf)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	application.Run()
}