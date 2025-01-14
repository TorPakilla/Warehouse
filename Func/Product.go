package Func

import (
	"Api/Models"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddProductWithInventory(db *gorm.DB, c *fiber.Ctx) error {
	type ProductRequest struct {
		ProductName     string  `json:"productname"`
		Description     string  `json:"description"`
		Type            string  `json:"type"`
		BrancheID       string  `json:"brancheid"`
		InitialQuantity uint32  `json:"initialquantity"`
		Price           float64 `json:"price"`
	}

	var req ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	// Validate fields
	if req.ProductName == "" || req.Type == "" || req.BrancheID == "" || req.InitialQuantity <= 0 || req.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All fields are required and must have valid values"})
	}

	// Create Product
	product := Models.Product{
		ProductName: req.ProductName,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}
	if err := db.Create(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create product: " + err.Error()})
	}

	// Determine Conversion Rate based on Type
	var conversRate *int
	if req.Type == "Pallet" {
		conversRate = new(int)
		*conversRate = 30
	} else if req.Type == "Box" {
		conversRate = new(int)
		*conversRate = 12
	}

	// Create Product Unit
	productUnit := Models.ProductUnit{
		ProductID:   product.ProductID,
		Type:        req.Type,
		ConversRate: conversRate,
	}
	if err := db.Create(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create product unit: " + err.Error()})
	}

	// Create Inventory
	inventory := Models.Inventory{
		ProductUnitID: productUnit.ProductUnitID,
		BrancheID:     req.BrancheID,
		Quantity:      req.InitialQuantity,
		Price:         req.Price,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create inventory: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Product, Product Unit, and Inventory created successfully",
		"product":     product,
		"productUnit": productUnit,
		"inventory":   inventory,
	})
}

func LookProduct(db *gorm.DB, c *fiber.Ctx) error {
	var products []Models.Product
	if err := db.Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find products: " + err.Error()})
	}
	return c.JSON(fiber.Map{"This": "Product", "Data": products})
}

func FindProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	return c.JSON(fiber.Map{"This": "Product", "Data": product})
}

func DeleteProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	if err := db.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete product: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": product})
}

func UpdateProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	productName := c.FormValue("productname")
	description := c.FormValue("description")
	productType := c.FormValue("type")

	if productName == "" || productType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product name and type are required"})
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		fileContent, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open uploaded file"})
		}
		defer fileContent.Close() // ปิดไฟล์หลังอ่านเสร็จ

		fileBytes, err := io.ReadAll(fileContent)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read uploaded file"})
		}

		product.Image = fileBytes // อัปเดตรูปภาพในฐานข้อมูล
	}

	product.ProductName = productName
	product.Description = description

	if err := db.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product: " + err.Error()})
	}

	// อัปเดต ProductUnit
	var productUnit Models.ProductUnit
	if err := db.Where("product_id = ?", id).First(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product unit not found"})
	}

	productUnit.Type = productType
	if productType == "Pallet" {
		productUnit.ConversRate = new(int)
		*productUnit.ConversRate = 30
	} else if productType == "Box" {
		productUnit.ConversRate = new(int)
		*productUnit.ConversRate = 12
	} else {
		productUnit.ConversRate = nil
	}

	if err := db.Save(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product unit: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"product": product, "productUnit": productUnit})
}

func ProductRouter(app fiber.Router, db *gorm.DB) {
	app.Use(func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "God" && role != "Manager" && role != "Stock" {
			return c.Next()
		}

		if role != "Account" && role != "Audit" {
			if c.Method() != "GET" {
				return c.Next()
			} else {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	})

	app.Post("/Product", func(c *fiber.Ctx) error {
		return AddProductWithInventory(db, c)
	})

	app.Get("/Product", func(c *fiber.Ctx) error {
		var products []Models.Product
		if err := db.Find(&products).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch products"})
		}
		return c.JSON(fiber.Map{"products": products})
	})

	app.Get("/Product/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var product Models.Product
		if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
		}
		return c.JSON(fiber.Map{"product": product})
	})

	app.Put("/Product/:id", func(c *fiber.Ctx) error {
		// Add update logic if needed
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Not implemented yet"})
	})

	app.Delete("/Product/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var product Models.Product
		if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
		}
		if err := db.Delete(&product).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete product"})
		}
		return c.JSON(fiber.Map{"message": "Product deleted successfully"})
	})
}
