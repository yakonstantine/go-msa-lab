package app

import (
	"fmt"
	"time"

	"github.com/yakonstantine/go-msa-lab/services/user-service/config"
)

func Run(cfg *config.Config) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Printf("%v: heartbeat\n", time.Now().Format(time.RFC3339))
	}
}
