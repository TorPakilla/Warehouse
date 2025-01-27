package Func

import (
	"Api/Models"

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
		JOIN public."Branches" b ON i.branch_id = b.branch_id
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

func InventoryRoutes(app *fiber.App, db *gorm.DB, posDB *gorm.DB) {
	app.Get("/BranchesWithInventory", func(c *fiber.Ctx) error {
		return GetBranchesWithInventory(db, posDB, c)
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
