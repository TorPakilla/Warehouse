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

// connectToDatabase establishes a connection to the database
func connectToDatabase(host string, port int, user, password, dbname string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Connect to Warehouse DB
	warehouseHost := os.Getenv("WAREHOUSE_DB_HOST")
	warehousePort, _ := strconv.Atoi(os.Getenv("WAREHOUSE_DB_PORT"))
	warehouseUser := os.Getenv("WAREHOUSE_DB_USER")
	warehousePassword := os.Getenv("WAREHOUSE_DB_PASSWORD")
	warehouseName := os.Getenv("WAREHOUSE_DB_NAME")

	db, err := connectToDatabase(warehouseHost, warehousePort, warehouseUser, warehousePassword, warehouseName)
	if err != nil {
		log.Fatalf("Failed to connect to Warehouse database: %v", err)
	}
	log.Println("Connected to Warehouse database!")

	// Connect to POS DB
	posHost := os.Getenv("POS_DB_HOST")
	posPort, _ := strconv.Atoi(os.Getenv("POS_DB_PORT"))
	posUser := os.Getenv("POS_DB_USER")
	posPassword := os.Getenv("POS_DB_PASSWORD")
	posName := os.Getenv("POS_DB_NAME")

	posDB, err := connectToDatabase(posHost, posPort, posUser, posPassword, posName)
	if err != nil {
		log.Fatalf("Failed to connect to POS database: %v", err)
	}
	log.Println("Connected to POS database!")

	go Func.StartSyncScheduler(db, posDB)
	// Initialize Fiber app
	app := fiber.New()

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000", // Allow frontend domain
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	})

	db.Migrator().DropTable(
		&Models.ShipmentItem{},
		&Models.Shipment{},
	// &Models.OrderItem{},
	// &Models.Order{},
	// &Models.ProductUnit{},
	// &Models.Inventory{},
	// &Models.Supplier{},
	// &Models.Product{},
	// &Models.Branches{},
	)

	// สร้างตารางใหม่ตามลำดับ
	if err := db.AutoMigrate(
		// &Models.Branches{},
		// &Models.Product{},
		// &Models.Supplier{},
		// &Models.Inventory{},
		// &Models.ProductUnit{},
		// &Models.Order{},
		// &Models.OrderItem{},
		// &Models.Employees{},
		&Models.Shipment{},
		&Models.ShipmentItem{},
	); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	// Authentication Routes
	app.Post("/login", Authentication.Login)

	// Protected Routes Example
	app.Use("/protected", func(c *fiber.Ctx) error {
		userName := c.Get("Authorization")
		if userName == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
		}
		return c.JSON(fiber.Map{"message": "You are authorized"})
	})

	// API Routes
	Func.EmployeesRoutes(app, db)
	Func.BranchRoutes(app, db, posDB)
	Func.ProductRouter(app, db)
	Func.InventoryRoutes(app, db, posDB)
	Func.SupplierRoutes(app, db)
	Func.OrderRoutes(app, db)
	Func.OrderItemRoutes(app, db)
	Func.ShipmentRoutes(app, db, posDB)
	Func.ShipmentItemRoutes(app, db)

	// Start server
	log.Println("Starting server on port 5050...")
	log.Fatal(app.Listen(":5050"))
}
