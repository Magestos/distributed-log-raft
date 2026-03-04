package main

import (
	"fmt"
	"log"

	"github.com/Magestos/distributed-log-raft/internal/config"
)

func main() {
	newcfg, err := config.Load("internal/config/config.yml")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(newcfg)
}
