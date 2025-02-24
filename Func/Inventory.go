package Func

import (
	"Api/Models"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// เพิ่มข้อมูล Inventory
func AddInventory(db *gorm.DB, c *fiber.Ctx) error {
	type InventoryRequest struct {
		ProductID string  `json:"product_id" validate:"required"`
		BranchID  string  `json:"branch_id" validate:"required"`
		Quantity  int     `json:"quantity" validate:"required,min=1"`
		Price     float64 `json:"price" validate:"required,min=0"`
	}

	var req InventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	if req.Quantity <= 0 || req.Price < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Quantity and Price must be greater than 0"})
	}

	inventory := Models.Inventory{
		ProductID: req.ProductID,
		BranchID:  req.BranchID,
		Quantity:  req.Quantity,
		Price:     req.Price,
	}

	if err := db.Create(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create inventory: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Inventory created successfully", "data": inventory})
}

// อัปเดตข้อมูล Inventory
func UpdateInventory(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var inventory Models.Inventory
	if err := db.Where("inventory_id = ?", id).First(&inventory).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory not found"})
	}

	type InventoryRequest struct {
		ProductID string  `json:"product_id"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
		BranchID  string  `json:"branch_id"`
	}

	var req InventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	if req.Quantity < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Quantity must be greater or equal to 0"})
	}

	inventory.ProductID = req.ProductID
	inventory.BranchID = req.BranchID
	inventory.Quantity = req.Quantity
	inventory.Price = req.Price

	if err := db.Save(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Inventory updated successfully", "data": inventory})
}

// ดึงข้อมูล Inventory ทั้งหมด
func LookInventory(db *gorm.DB, c *fiber.Ctx) error {
	var inventories []Models.Inventory
	if err := db.Find(&inventories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot fetch inventory data: " + err.Error()})
	}

	return c.JSON(fiber.Map{"data": inventories})
}

// ค้นหาข้อมูล Inventory
func FindInventory(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var inventory Models.Inventory
	if err := db.Where("inventory_id = ?", id).First(&inventory).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory not found"})
	}

	return c.JSON(fiber.Map{"data": inventory})
}

// ลบข้อมูล Inventory
func DeleteInventory(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var inventory Models.Inventory
	if err := db.Where("inventory_id = ?", id).First(&inventory).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Inventory not found"})
	}
	if err := db.Delete(&inventory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete inventory: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Inventory deleted successfully"})
}

// ดึงข้อมูล Inventory ตาม Branch ID
func GetInventoriesByBranch(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	branchID := c.Query("branch_id")
	if branchID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "branch_id query parameter is required"})
	}

	var inventories []Models.Inventory

	// Query ข้อมูลใน Warehouse
	if err := db.Where("branch_id = ?", branchID).Find(&inventories).Error; err == nil && len(inventories) > 0 {
		return c.JSON(fiber.Map{"inventories": inventories})
	}

	// Query ข้อมูลใน POS
	if err := posDB.Where("branch_id = ?", branchID).Find(&inventories).Error; err == nil && len(inventories) > 0 {
		return c.JSON(fiber.Map{"inventories": inventories})
	}

	return c.JSON(fiber.Map{"inventories": []interface{}{}})
}

// ดึง Branches ที่มี Inventory
func GetBranchesWithInventory(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	var warehouseBranches []struct {
		BranchID   string `json:"branch_id"`
		BranchName string `json:"branch_name"`
	}

	var posBranches []struct {
		BranchID   string `json:"branch_id"`
		BranchName string `json:"branch_name"`
	}

	// ดึงข้อมูลจาก Warehouse
	warehouseErr := db.Raw(`
    SELECT DISTINCT i.branch_id AS branch_id, b.b_name AS branch_name
    FROM public."Inventory" i
    JOIN public."Branches" b ON i.branch_id::UUID = b.branch_id
    WHERE i.quantity > 0
`).Scan(&warehouseBranches).Error

	// ดึงข้อมูลจาก POS
	posErr := posDB.Raw(`
		SELECT DISTINCT i.branch_id AS branch_id, b.b_name AS branch_name
		FROM public."Inventory" i
		JOIN public."Branches" b ON i.branch_id = b.branch_id
		WHERE i.quantity > 0
	`).Scan(&posBranches).Error

	// ตรวจสอบข้อผิดพลาด
	if warehouseErr != nil && posErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch branches with inventory",
			"details": "Warehouse Error: " + warehouseErr.Error() + ", POS Error: " + posErr.Error(),
		})
	}

	// แยกข้อมูลสำหรับ frontend
	return c.JSON(fiber.Map{
		"warehouse_branches": warehouseBranches,
		"pos_branches":       posBranches,
	})
}

func GetInventorySummary(db *gorm.DB, c *fiber.Ctx) error {
	var inventoryData []struct {
		ProductID   string `json:"product_id"`
		ProductName string `json:"product_name"`
		Description string `json:"description"`
		Quantity    int    `json:"quantity"`
	}

	// Query รวมข้อมูล
	err := db.Raw(`
		SELECT p.product_id, p.product_name, p.description, SUM(i.quantity) as quantity
		FROM public."Inventory" i
		JOIN public."Product" p ON i.product_id = p.product_id
		GROUP BY p.product_id, p.product_name, p.description
	`).Scan(&inventoryData).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch inventory summary",
		})
	}

	return c.JSON(fiber.Map{"inventory_summary": inventoryData})
}

type InventoryCategory struct {
	Category      string          `json:"category"`
	TotalQuantity int             `json:"total_quantity"`
	Details       json.RawMessage `json:"details"` // ใช้ json.RawMessage เพื่อเก็บ JSON ดิบ
}

func GetInventoryByCategory(db *gorm.DB, c *fiber.Ctx) error {
	var categories []struct {
		Category      string          `json:"category"`
		TotalQuantity int             `json:"total_quantity"`
		Details       json.RawMessage `json:"details"`
	}

	err := db.Raw(`
    SELECT 
        p.description AS category,
        COALESCE(SUM(i.quantity), 0) AS total_quantity,
        JSON_AGG(
            JSON_BUILD_OBJECT('product_name', p.product_name, 'quantity', i.quantity)
        ) AS details
    FROM public."Inventory" i
    JOIN public."Product" p ON i.product_id::UUID = p.product_id
    GROUP BY p.description
`).Scan(&categories).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch inventory by category: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{"categories": categories})
}

var lowStockCache []struct {
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	BranchName  string    `json:"branch_name"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var lastCacheTime time.Time

func GetPosLowStock(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	if time.Since(lastCacheTime) < 30*time.Second {
		log.Println("🔄 Using cached POS Low Stock data")
		return c.JSON(fiber.Map{"low_stock_items": lowStockCache})
	}

	log.Println("Fetching POS Low Stock Items from Database...")

	err := posDB.Raw(`
        SELECT 
            p.product_id, 
            p.product_name, 
            c.category_name AS category, 
            i.quantity, 
            b.b_name AS branch_name,
			i.updated_at AT TIME ZONE 'UTC' AS updated_at
        FROM public."Inventory" i
        JOIN public."Products" p ON i.product_id = p.product_id
        JOIN public."Category" c ON p.category_id = c.category_id  
        JOIN public."Branches" b ON i.branch_id = b.branch_id  
        WHERE i.quantity < 1000
        ORDER BY i.updated_at DESC 
    `).Scan(&lowStockCache).Error

	if err != nil {
		log.Println("❌ Error fetching POS low stock items:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch POS low stock items: " + err.Error(),
		})
	}

	lastCacheTime = time.Now() // ✅ อัปเดต cache time
	return c.JSON(fiber.Map{"low_stock_items": lowStockCache})
}

func GetFilteredCategories(db *gorm.DB, posDB *gorm.DB, c *fiber.Ctx) error {
	fromBranch := c.Query("fromBranch")
	toBranch := c.Query("toBranch")

	if fromBranch == "" || toBranch == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "fromBranch and toBranch are required"})
	}

	var categories []string

	// Query ดึงหมวดหมู่สินค้าที่มีอยู่ในทั้ง Warehouse และ POS ที่เลือก
	query := `
        SELECT DISTINCT p.description 
        FROM public."Inventory" i
        JOIN public."Product" p ON i.product_id = p.product_id
        WHERE i.branch_id IN (?, ?)
    `

	err := db.Raw(query, fromBranch, toBranch).Scan(&categories).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"categories": categories})
}

func GetProductsByCategoryAndBranch(db *gorm.DB, c *fiber.Ctx) error {
	branchID := c.Query("branchId")
	category := c.Query("category")

	if branchID == "" || category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "branchId and category are required"})
	}

	var products []struct {
		ProductID   string `json:"product_id"`
		ProductName string `json:"product_name"`
	}

	err := db.Raw(`
        SELECT DISTINCT p.product_id, p.product_name
        FROM public."Inventory" i
        JOIN public."Product" p ON i.product_id = p.product_id
        WHERE i.branch_id = ? AND LOWER(p.description) = LOWER(?)
    `, branchID, category).Scan(&products).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"products": products})
}

func GetMatchingProductsInPOS(posDB *gorm.DB, c *fiber.Ctx) error {
	branchID := c.Query("branchId")
	productName := c.Query("productName")

	if branchID == "" || productName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "branchId and productName are required"})
	}

	var products []struct {
		ProductID   string `json:"product_id"`
		ProductName string `json:"product_name"`
		InventoryID string `json:"inventory_id"` // ✅ เพิ่มตรงนี้
	}

	err := posDB.Raw(`
		SELECT DISTINCT p.product_id, p.product_name, i.inventory_id
		FROM public."Inventory" i
		JOIN public."Products" p ON i.product_id = p.product_id
		WHERE i.branch_id = ? AND p.product_name ILIKE ?
	`, branchID, "%"+productName+"%").Scan(&products).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"products": products})
}

func InventoryRoutes(app *fiber.App, db *gorm.DB, posDB *gorm.DB) {

	app.Get("/GetProductsByCategoryAndBranch", func(c *fiber.Ctx) error {
		return GetProductsByCategoryAndBranch(db, c)
	})

	app.Get("/GetMatchingProductsInPOS", func(c *fiber.Ctx) error {
		return GetMatchingProductsInPOS(posDB, c)
	})

	app.Get("/GetFilteredCategories", func(c *fiber.Ctx) error {
		return GetFilteredCategories(db, posDB, c)
	})

	app.Get("/BranchesWithInventory", func(c *fiber.Ctx) error {
		return GetBranchesWithInventory(db, posDB, c)
	})

	app.Get("/GetPosLowStock", func(c *fiber.Ctx) error {
		return GetPosLowStock(db, posDB, c)
	})

	app.Get("/inventory-summary", func(c *fiber.Ctx) error {
		return GetInventorySummary(db, c)
	})

	app.Get("/inventory-by-category", func(c *fiber.Ctx) error {
		return GetInventoryByCategory(db, c)
	})

	app.Get("/InventoriesByBranch", func(c *fiber.Ctx) error {
		return GetInventoriesByBranch(db, posDB, c)
	})

	app.Get("/Inventory", func(c *fiber.Ctx) error {
		return LookInventory(db, c)
	})

	app.Get("/Inventory/:id", func(c *fiber.Ctx) error {
		return FindInventory(db, c)
	})

	app.Post("/Inventory", func(c *fiber.Ctx) error {
		return AddInventory(db, c)
	})

	app.Put("/Inventory/:id", func(c *fiber.Ctx) error {
		return UpdateInventory(db, c)
	})

	app.Delete("/Inventory/:id", func(c *fiber.Ctx) error {
		return DeleteInventory(db, c)
	})
}
