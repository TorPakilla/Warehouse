package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"Api/Authentication"
	"Api/Func"
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Protected(c *fiber.Ctx) error {
	UserName := c.Get("Authorization")
	if UserName == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"message": "You are authorized"})
}

func connectToDatabase(host string, port int, user, password, dbname string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	warehouseHost := os.Getenv("WAREHOUSE_DB_HOST")
	warehousePort, _ := strconv.Atoi(os.Getenv("WAREHOUSE_DB_PORT"))
	warehouseUser := os.Getenv("WAREHOUSE_DB_USER")
	warehousePassword := os.Getenv("WAREHOUSE_DB_PASSWORD")
	warehouseName := os.Getenv("WAREHOUSE_DB_NAME")

	posHost := os.Getenv("POS_DB_HOST")
	posPort, _ := strconv.Atoi(os.Getenv("POS_DB_PORT"))
	posUser := os.Getenv("POS_DB_USER")
	posPassword := os.Getenv("POS_DB_PASSWORD")
	posName := os.Getenv("POS_DB_NAME")

	db, err := connectToDatabase(warehouseHost, warehousePort, warehouseUser, warehousePassword, warehouseName)
	if err != nil {
		log.Fatalf("Failed to connect to Warehouse database: %v", err)
	}
	fmt.Println("Connected to Warehouse database!")

	posDB, err := connectToDatabase(posHost, posPort, posUser, posPassword, posName)
	if err != nil {
		log.Fatalf("Failed to connect to POS database: %v", err)
	}
	fmt.Println("Connected to POS database!")

	// ลบตารางเก่า
	db.Migrator().DropTable(
		&Models.ShipmentItem{},
		&Models.Shipment{},
		&Models.OrderItem{},
		&Models.Order{},
	// &Models.ProductUnit{},
	// &Models.Inventory{},
	// &Models.Supplier{},
	// &Models.Product{},
	// &Models.Branches{},
	)

	// เรียก AutoMigrate ใหม่ในลำดับที่ถูกต้อง
	if err := db.AutoMigrate(
		// &Models.Branches{},
		// &Models.Product{},
		// &Models.Supplier{},
		// &Models.Inventory{},
		// &Models.ProductUnit{},
		&Models.Order{},
		&Models.OrderItem{},
		// // &Models.Employees{},
		&Models.Shipment{},
		&Models.ShipmentItem{},
	); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	log.Println("Database migration completed successfully.")

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000", // อนุญาต domain frontend
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Post("/login", Authentication.Login)

	app.Use("/protected", Protected)

	Func.EmployeesRoutes(app, db)
	Func.BranchRoutes(app, db, posDB)
	Func.ProductRouter(app, db)
	Func.InventoryRoutes(app, db, posDB)
	Func.SupplierRoutes(app, db)
	Func.OrderRoutes(app, db)
	Func.OrderItemRoutes(app, db)
	Func.ShipmentRoutes(app, db, posDB)
	Func.ShipmentItemRoutes(app, db)

	log.Fatal(app.Listen(":5050"))
}
