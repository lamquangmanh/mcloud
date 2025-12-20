package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"mcloud/internal/config"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	nodeName, _ := os.Hostname()

	req := map[string]string{
		"Node": nodeName,
	}

	body, _ := json.Marshal(req)

	_, err = http.Post(
		cfg.Agent.ManagerURL+"/register",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		log.Fatal(err)
	}

}
