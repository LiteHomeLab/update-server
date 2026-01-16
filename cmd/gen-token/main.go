package main

import (
	"fmt"
	"log"

	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/service"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect database
	db, err := database.NewGORM(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Create token service
	tokenSvc := service.NewTokenService(db)

	// Generate admin token
	token, tokenValue, err := tokenSvc.GenerateToken("", "admin", "system")
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	fmt.Printf("Admin Token: %s\n", tokenValue)
	fmt.Printf("Token ID: %s\n", token.TokenID)
}
