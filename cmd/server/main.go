package main

import (
	"log"
	"os"

	"gorepositorytest/internal/middleware"
	"gorepositorytest/internal/models"
	"gorepositorytest/internal/repository"
	"gorepositorytest/internal/routes"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	middleware.RegisterMiddlewares(r)

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	productRepo := repository.NewPostgresProductRepository(db)
	orderRepo := repository.NewPostgresOrderRepository(db)

	routes.SetupProductRoutes(r, productRepo)
	routes.SetupOrderRoutes(r, orderRepo, productRepo)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func initDatabase() (*gorm.DB, error) {
	dsn := "host=localhost user=devuser password=devpassword dbname=devdb port=5432 sslmode=disable"
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		dsn = envDSN
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
