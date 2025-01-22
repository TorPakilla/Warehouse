package Func

import (
	"Api/Models"
	"io"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddProductWithInventory(db *gorm.DB, c *fiber.Ctx) error {
	type ProductRequest struct {
		ProductName     string
		Description     string
		Type            string
		BranchID        string
		InitialQuantity int
		Price           float64
		Image           []byte
	}

	req := ProductRequest{}
	req.ProductName = c.FormValue("product_name")
	req.Description = c.FormValue("description")
	req.Type = c.FormValue("type")
	req.BranchID = c.FormValue("branch_id")

	// แปลง initial_quantity
	initialQuantity, err := strconv.Atoi(c.FormValue("initial_quantity"))
	if err != nil || initialQuantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid initial_quantity"})
	}
	req.InitialQuantity = initialQuantity

	// แปลง price
	price, err := strconv.ParseFloat(c.FormValue("price"), 64)
	if err != nil || price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid price"})
	}
	req.Price = price

	// Handle file upload
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		fileContent, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open uploaded file"})
		}
		defer fileContent.Close()

		req.Image, err = io.ReadAll(fileContent)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read uploaded file"})
		}
	}

	// 1. สร้าง Product
	product := Models.Product{
		ProductName: req.ProductName,
		Description: req.Description,
		Image:       req.Image,
		CreatedAt:   time.Now(),
	}

	// บันทึก Product ลงในฐานข้อมูล
	if err := db.Create(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create product"})
	}

	// ดึง Product ที่บันทึกลงฐานข้อมูล
	if err := db.Last(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch created product"})
	}

	// 2. กำหนด ConversRate ตามประเภทของสินค้า
	var conversRate int
	switch req.Type {
	case "Pallet":
		conversRate = 12
	case "Box":
		conversRate = 6
	case "Pieces":
		conversRate = 1
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid type"})
	}

	// 3. สร้าง ProductUnit ที่เชื่อมโยงกับ Product ที่สร้างขึ้น
	productUnit := Models.ProductUnit{
		ProductID:       product.ProductID,
		Type:            req.Type,
		InitialQuantity: req.InitialQuantity,
		ConversRate:     conversRate,
		CreatedAt:       time.Now(),
	}

	// บันทึก ProductUnit ลงในฐานข้อมูล
	if err := db.Create(&productUnit).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create product unit"})
	}

	// 4. คำนวณและสร้าง Inventory ที่เชื่อมโยงกับ Product และ ProductUnit
	calculatedQuantity := req.InitialQuantity * conversRate
	inventory := Models.Inventory{
		ProductID: product.ProductID,
		BranchID:  req.BranchID, // เพิ่ม branch_id ใน Inventory
		Quantity:  calculatedQuantity,
		Price:     req.Price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// บันทึก Inventory ลงในฐานข้อมูล
	if err := db.Create(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create inventory"})
	}

	// 5. ส่งผลลัพธ์สำเร็จ
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Product, Product Unit, and Inventory created successfully",
		"product":     product,
		"productUnit": productUnit,
		"inventory":   inventory,
	})
}

func UpdateProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	productName := c.FormValue("product_name")
	description := c.FormValue("description")

	if productName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product name is required"})
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		fileContent, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open uploaded file"})
		}
		defer fileContent.Close()

		fileBytes, err := io.ReadAll(fileContent)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read uploaded file"})
		}

		product.Image = fileBytes
	}

	product.ProductName = productName
	product.Description = description

	if err := db.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"product": product})
}

func LookProducts(db *gorm.DB, c *fiber.Ctx) error {
	var products []Models.Product
	if err := db.Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch products"})
	}
	return c.JSON(fiber.Map{"products": products})
}

func DeleteProduct(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var product Models.Product
	if err := db.Where("product_id = ?", id).First(&product).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	if err := db.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete product"})
	}

	return c.JSON(fiber.Map{"message": "Product deleted successfully"})
}

func ProductRouter(app fiber.Router, db *gorm.DB) {
	app.Post("/Product", func(c *fiber.Ctx) error {
		return AddProductWithInventory(db, c)
	})
	app.Get("/Product", func(c *fiber.Ctx) error {
		return LookProducts(db, c)
	})
	app.Put("/Product/:id", func(c *fiber.Ctx) error {
		return UpdateProduct(db, c)
	})
	app.Delete("/Product/:id", func(c *fiber.Ctx) error {
		return DeleteProduct(db, c)
	})
}
