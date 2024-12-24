// filepath: /c:/Users/dY470g3/Desktop/ProjectWork/Warehouse/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"Api/Database"
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SupplyRoutes(app *fiber.App, db *gorm.DB) {

	app.Post("/Supply", func(c *fiber.Ctx) error {
		return Database.AddOrderSupply(db, c)
	})

	app.Get("/Supply", func(c *fiber.Ctx) error {
		return Database.LookOrderSupply(db, c)
	})

	app.Get("/Supply/:id", func(c *fiber.Ctx) error {
		return Database.LookOrderSupplyById(db, c)
	})

	app.Delete("/Supply", func(c *fiber.Ctx) error {
		return Database.DeleteOrderSupply(db, c)
	})

	app.Put("/Supply", func(c *fiber.Ctx) error {
		return Database.UpdateOrderSupply(db, c)
	})
}

func SupplierRoutes(app *fiber.App, db *gorm.DB) {

	app.Post("/Supplier", func(c *fiber.Ctx) error {
		return Database.AddSupplier(db, c)
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&Models.OrderSupply{}, &Models.Supplier{})

	app := fiber.New()
	SupplyRoutes(app, db)
	SupplierRoutes(app, db)

	log.Fatal(app.Listen(":5050"))
}
