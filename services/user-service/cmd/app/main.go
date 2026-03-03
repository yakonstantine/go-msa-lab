package main

import (
	"log"

	"github.com/yakonstantine/go-msa-lab/services/user-service/config"
	"github.com/yakonstantine/go-msa-lab/services/user-service/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
