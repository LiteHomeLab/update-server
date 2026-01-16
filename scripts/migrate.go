package main

import (
	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	loggerCfg := logger.Config{
		Level:      cfg.Logger.Level,
		Output:     cfg.Logger.Output,
		FilePath:   cfg.Logger.FilePath,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}
	if err := logger.Init(loggerCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := database.NewGORM(cfg.Database.Path)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting migration...")

	// Step 1: Create docufiller program record
	fmt.Println("\n[Step 1] Creating docufiller program record...")

	// Check if program already exists
	var existingProgram models.Program
	err = db.Where("program_id = ?", "docufiller").First(&existingProgram).Error
	if err == nil {
		fmt.Printf("Program 'docufiller' already exists (ID: %d)\n", existingProgram.ID)
	} else {
		program := &models.Program{
			ProgramID:   "docufiller",
			Name:        "DocuFiller",
			Description: "文档填充工具",
			IsActive:    true,
		}
		if err := db.Create(program).Error; err != nil {
			fmt.Printf("Failed to create program: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Created docufiller program")
	}

	// Step 2: Update existing version records
	fmt.Println("\n[Step 2] Updating existing version records...")

	// Count versions to be updated
	var count int64
	db.Model(&models.Version{}).
		Where("program_id IS NULL OR program_id = ''").
		Count(&count)

	if count > 0 {
		result := db.Model(&models.Version{}).
			Where("program_id IS NULL OR program_id = ''").
			Update("program_id", "docufiller")
		if result.Error != nil {
			fmt.Printf("Failed to update versions: %v\n", result.Error)
			os.Exit(1)
		}
		fmt.Printf("✓ Updated %d version records\n", result.RowsAffected)
	} else {
		fmt.Println("✓ No version records need updating (all have program_id)")
	}

	// Step 3: Generate initial tokens
	fmt.Println("\n[Step 3] Generating initial tokens...")

	tokenSvc := service.NewTokenService(db)

	// Generate upload token
	var uploadValue string
	err = db.Where("program_id = ? AND token_type = ?", "docufiller", "upload").First(&models.Token{}).Error
	if err == nil {
		fmt.Println("Upload token already exists, skipping generation")
	} else {
		_, uploadValue, err = tokenSvc.GenerateToken("docufiller", "upload", "migration")
		if err != nil {
			fmt.Printf("Failed to generate upload token: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated upload token: %s\n", uploadValue)
	}

	// Generate download token
	var downloadValue string
	err = db.Where("program_id = ? AND token_type = ?", "docufiller", "download").First(&models.Token{}).Error
	if err == nil {
		fmt.Println("Download token already exists, skipping generation")
	} else {
		_, downloadValue, err = tokenSvc.GenerateToken("docufiller", "download", "migration")
		if err != nil {
			fmt.Printf("Failed to generate download token: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated download token: %s\n", downloadValue)
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("Migration completed successfully!")
	fmt.Println(strings.Repeat("-", 50))

	if uploadValue != "" || downloadValue != "" {
		fmt.Println("\nIMPORTANT: Save these tokens securely:")
		if uploadValue != "" {
			fmt.Printf("  Upload Token:   %s\n", uploadValue)
		}
		if downloadValue != "" {
			fmt.Printf("  Download Token: %s\n", downloadValue)
		}
		fmt.Println("\nStore these tokens in a secure location.")
		fmt.Println("You will need them to access the API endpoints.")
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Run migrate-storage.sh to migrate package files")
	fmt.Println("2. Test the API with the generated tokens")
	fmt.Println("3. Update your client applications to use the new tokens")
}
