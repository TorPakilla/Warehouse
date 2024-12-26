package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"Api/Func"
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

	// db.Migrator().DropTable(
	// 	&Models.Employees{},
	// 	&Models.Branches{},
	// 	&Models.Product{},
	// 	&Models.ProductUnit{},
	// 	&Models.Inventory{},
	// 	&Models.Supplier{},
	// 	&Models.Order{},
	// 	&Models.OrderItem{},
	// 	&Models.Shipment{},
	// 	&Models.ShipmentItem{},
	// )

	db.AutoMigrate(&Models.Employees{},
		&Models.Branches{},
		&Models.Product{},
		&Models.ProductUnit{},
		&Models.Inventory{},
		&Models.Supplier{},
		&Models.Order{},
		&Models.OrderItem{},
		&Models.Shipment{},
		&Models.ShipmentItem{})

	app := fiber.New()

	Func.EmployeesRoutes(app, db)
	Func.BranchesRoutes(app, db)
	Func.ProductRouter(app, db)
	Func.ProductUnitRouter(app, db)
	Func.InventoryRoutes(app, db)
	Func.SupplierRoutes(app, db)
	Func.OrderRoutes(app, db)
	Func.OrderItemRoutes(app, db)
	Func.ShipmentRoutes(app, db)
	Func.ShipmentItemRoutes(app, db)

	log.Fatal(app.Listen(":5050"))
}
