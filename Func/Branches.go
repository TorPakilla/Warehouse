package Func

import (
	"Api/Models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AddBranches adds a new branch.
func AddBranches(db *gorm.DB, c *fiber.Ctx) error {
	type Request struct {
		BName    string `json:"b_name" validate:"required"` // เปลี่ยนจาก bname เป็น b_name
		Location string `json:"location" validate:"required"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	if req.BName == "" || req.Location == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "BName and Location are required"})
	}

	branch := Models.Branches{
		BName:    req.BName,
		Location: req.Location,
	}

	if err := db.Create(&branch).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create branch"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"new_branch": branch})
}

// UpdateBranches updates a branch by its ID
func UpdateBranches(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Branch ID is required"})
	}

	fmt.Println("Updating Branch ID:", id) // Debug log
	var branch Models.Branches
	if err := db.Where("branch_id = ?", id).First(&branch).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branch not found"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	// Update fields in the branch
	if err := db.Model(&branch).Updates(body).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update branch"})
	}

	return c.JSON(fiber.Map{"message": "Branch updated successfully", "branch": branch})
}

// LookBranch retrieves all branches.
func LookBranch(db *gorm.DB, c *fiber.Ctx) error {
	var branches []Models.Branches
	if err := db.Find(&branches).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch branches"})
	}
	return c.JSON(fiber.Map{"branches": branches})
}

// FindBranches retrieves a branch by its ID.
func FindBranches(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var branch Models.Branches
	if err := db.Where("branch_id = ?", id).First(&branch).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branch not found"})
	}
	return c.JSON(fiber.Map{"branch": branch})
}

// DeleteBranches deletes a branch by its ID.
func DeleteBranches(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var branch Models.Branches
	if err := db.Where("branch_id = ?", id).First(&branch).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branch not found"})
	}

	if err := db.Delete(&branch).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete branch: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Branch deleted successfully"})
}

// GetWarehouseInventory retrieves inventory for a specific branch ID.
func GetWarehouseInventory(db *gorm.DB, c *fiber.Ctx) error {
	branchID := c.Query("branchId")
	if branchID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Branch ID is required"})
	}

	var inventoryItems []Models.Inventory
	if err := db.Where("branch_id = ?", branchID).Find(&inventoryItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch inventory items"})
	}

	return c.JSON(fiber.Map{"inventoryItems": inventoryItems})
}

func GetPOSInventory(posDB *gorm.DB, c *fiber.Ctx) error {
	branchID := c.Query("branchId")
	if branchID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Branch ID is required"})
	}

	var inventoryItems []struct {
		InventoryID string `json:"inventory_id"`
		ProductID   string `json:"product_id"` // ใช้ product_id หลังจากการเปลี่ยนชื่อคอลัมน์
		BranchID    string `json:"branch_id"`
		Quantity    int    `json:"quantity"`
		UpdatedAt   string `json:"updated_at"`
	}

	if err := posDB.Table("Inventory").
		Select("inventory_id, product_id, branch_id, quantity, updated_at").
		Where("branch_id = ?", branchID).
		Find(&inventoryItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch POS inventory", "details": err.Error()})
	}

	return c.JSON(fiber.Map{"inventoryItems": inventoryItems})
}

func GetPOSBranches(posDB *gorm.DB, c *fiber.Ctx) error {
	var branches []struct {
		BranchID string `json:"branch_id"`
		BName    string `json:"b_name"`
		Location string `json:"location"`
	}

	// ดึงข้อมูลทั้งหมดจากตาราง Branches ของฐานข้อมูล POS
	if err := posDB.Table("Branches").Select("branch_id, b_name, location").Find(&branches).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch POS branches", "details": err.Error()})
	}

	// ตรวจสอบว่ามีข้อมูลหรือไม่
	if len(branches) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "No POS branches found"})
	}

	return c.JSON(fiber.Map{"branches": branches})
}

// GetWarehouseBranches retrieves Warehouse branches only.
func GetWarehouseBranches(db *gorm.DB, c *fiber.Ctx) error {
	var branches []struct {
		BranchID string `json:"branch_id"`
		BName    string `json:"b_name"`
		Location string `json:"location"`
	}

	// ดึงข้อมูลทั้งหมดจากตาราง Branches ของฐานข้อมูล Warehouse
	if err := db.Table("Branches").Select("branch_id, b_name, location").Find(&branches).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch Warehouse branches", "details": err.Error()})
	}

	// ตรวจสอบว่ามีข้อมูลหรือไม่
	if len(branches) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "No Warehouse branches found"})
	}

	return c.JSON(fiber.Map{"branches": branches})
}

func BranchRoutes(app *fiber.App, db *gorm.DB, posDB *gorm.DB) {
	app.Get("/WarehouseBranches", func(c *fiber.Ctx) error {
		return GetWarehouseBranches(db, c)
	})

	app.Get("/POSBranches", func(c *fiber.Ctx) error {
		return GetPOSBranches(posDB, c)
	})

	app.Get("/POSInventory", func(c *fiber.Ctx) error {
		return GetPOSInventory(posDB, c)
	})

	app.Get("/WarehouseInventory", func(c *fiber.Ctx) error {
		return GetWarehouseInventory(db, c)
	})

	app.Post("/Branches", func(c *fiber.Ctx) error {
		return AddBranches(db, c)
	})

	app.Get("/Branches", func(c *fiber.Ctx) error {
		return LookBranch(db, c)
	})

	app.Get("/Branches/:id", func(c *fiber.Ctx) error {
		return FindBranches(db, c)
	})

	app.Delete("/Branches/:id", func(c *fiber.Ctx) error {
		return DeleteBranches(db, c)
	})

	app.Put("/Branches/:id", func(c *fiber.Ctx) error {
		return UpdateBranches(db, c)
	})
}
