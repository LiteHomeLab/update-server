package main

import (
	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/models"
	"fmt"
	"log"
)

func main() {
	db, err := database.NewGORM("./data/versions.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 硬删除所有被软删除的版本记录
	result := db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Version{})
	if result.Error != nil {
		log.Fatalf("Failed to cleanup soft-deleted records: %v", result.Error)
	}

	fmt.Printf("Successfully cleaned up %d soft-deleted version records\n", result.RowsAffected)
}
