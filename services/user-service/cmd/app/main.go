package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Printf("%v: heartbeat\n", time.Now().Format(time.RFC3339))
	}
}
