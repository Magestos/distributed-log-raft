package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/Magestos/distributed-log-raft/internal/config"
)

func main() {
	configPath := flag.String("config", "", "path to node YAML config")
	flag.Parse()

	if strings.TrimSpace(*configPath) == "" {
		log.Fatal("config path is required, use -config")
	}

	newcfg, err := config.Load(*configPath)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(newcfg)
}
