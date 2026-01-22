package main

import (
	"fmt"
	"log"

	"docufiller-update-server/internal/database"
	"docufiller-update-server/internal/models"
)

func main() {
	// 连接数据库
	db, err := database.NewGORM("./data/versions.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 查找现有的管理员用户
	var admin models.AdminUser
	result := db.Where("username = ?", "admin").First(&admin)

	if result.Error == nil {
		// 管理员用户已存在，重置密码
		if err := admin.SetPassword("admin123"); err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}
		if err := db.Save(&admin).Error; err != nil {
			log.Fatalf("Failed to update admin user: %v", err)
		}
		fmt.Println("Admin password reset successfully!")
		fmt.Println("Username: admin")
		fmt.Println("Password: admin123")
	} else {
		// 创建新的管理员用户
		admin := models.AdminUser{
			Username: "admin",
		}
		if err := admin.SetPassword("admin123"); err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		fmt.Println("Admin user created successfully!")
		fmt.Println("Username: admin")
		fmt.Println("Password: admin123")
	}
}
