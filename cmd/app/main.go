// Package main implements a Telegram bot task tracker application.
//
// Example usage:
//
//	go run cmd/app/main.go
package main

import (
	"fmt"
	"log"

	"github.com/pvxdv/self_improver/internal/config"
	"github.com/pvxdv/self_improver/internal/config/loader/env"
)

func main() {
	cfgLoader := env.NewLoader()
	cfg, err := config.LoadAndValidate(cfgLoader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded config:\n%+v\n", cfg)

	//TODO logger

	//TODO storage

	//TODO http-server

	//TODO swagger

	//TODO graceful shutdown

	//TODO auth

	//TODO tbot
}
