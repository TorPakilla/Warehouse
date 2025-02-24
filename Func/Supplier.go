package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ✅ ฟังก์ชันเพิ่ม Supplier พร้อมบันทึกลง ProductSupplier
func AddSupplier(db *gorm.DB, c *fiber.Ctx) error {
	// รับข้อมูลจาก JSON Request
	type SupplierRequest struct {
		Name        string  `json:"name"`
		PricePallet float64 `json:"pricepallet"`
		ProductID   string  `json:"productid"`
	}

	var req SupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	// ✅ ตรวจสอบค่าที่จำเป็นต้องมี
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Supplier name is required"})
	}
	if req.PricePallet <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Price pallet must be greater than zero"})
	}

	// ✅ สร้าง UUID ใหม่ให้ Supplier (ใช้ค่าเดียวกันทุกที่)
	supplierUUID := uuid.New()

	// ✅ ตรวจสอบ `ProductID` ถ้ามีค่าต้องเป็น UUID ที่ถูกต้อง
	var productUUID *uuid.UUID
	if req.ProductID != "" {
		parsedUUID, err := uuid.Parse(req.ProductID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product UUID format"})
		}
		productUUID = &parsedUUID
	}

	// ✅ สร้าง Supplier ในฐานข้อมูล
	supplier := Models.Supplier{
		SupplierID:  supplierUUID.String(), // ✅ ใช้ UUID ที่เพิ่งสร้าง
		Name:        req.Name,
		PricePallet: req.PricePallet,
	}

	// ✅ ถ้ามี `ProductID` ให้ใส่ค่าลงไป
	if productUUID != nil {
		supplier.ProductID = productUUID.String()
	}

	if err := db.Create(&supplier).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create supplier: " + err.Error()})
	}

	// ✅ ถ้ามี `ProductID` ให้บันทึกลง `ProductSupplier` (ใช้ `supplierUUID` ที่ถูกต้อง)
	if productUUID != nil {
		productSupplier := Models.ProductSupplier{
			SupplierID: supplierUUID, // ✅ ใช้ UUID ที่สร้างตอนแรก
			ProductID:  *productUUID, // ✅ UUID ของ Product
		}

		if err := db.Create(&productSupplier).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to link product with supplier: " + err.Error()})
		}
	}

	// ✅ ตอบกลับข้อมูลที่สร้างสำเร็จ
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Supplier created successfully",
		"data": fiber.Map{
			"supplier_id": supplierUUID.String(), // ✅ ใช้ค่า supplierUUID ที่ถูกต้อง
			"name":        supplier.Name,
			"pricepallet": supplier.PricePallet,
		},
	})
}

// ดูข้อมูล Supplier
func LookSuppliers(db *gorm.DB, c *fiber.Ctx) error {
	var suppliers []Models.Supplier
	if err := db.Find(&suppliers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch suppliers: " + err.Error()})
	}
	return c.JSON(fiber.Map{"data": suppliers})
}

// หาข้อมูล Supplier ตาม ID
func FindSupplier(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var supplier Models.Supplier
	if err := db.Where("supplier_id = ?", id).First(&supplier).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}
	return c.JSON(fiber.Map{"data": supplier})
}

// อัพเดตข้อมูล Supplier
func UpdateSupplier(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var supplier Models.Supplier
	if err := db.Where("supplier_id = ?", id).First(&supplier).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}

	type SupplierRequest struct {
		Name        string  `json:"name"`
		PricePallet float64 `json:"pricepallet"`
		ProductID   string  `gorm:"foreignKey:ProductID" json:"productid"`
	}

	body := make(map[string]interface{})
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	allowedFields := map[string]bool{
		"name":        true,
		"pricepallet": true,
		"productid":   true,
	}

	for key := range body {
		if !allowedFields[key] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid field: " + key})
		}
	}

	var req SupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	supplier.Name = req.Name
	supplier.PricePallet = req.PricePallet
	supplier.ProductID = req.ProductID

	if err := db.Save(&supplier).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update supplier: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier updated successfully", "data": supplier})
}

// ลบข้อมูล Supplier ตาม ID
func DeleteSupplier(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var supplier Models.Supplier
	if err := db.Where("supplier_id = ?", id).First(&supplier).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Supplier not found"})
	}
	if err := db.Delete(&supplier).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete supplier: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Supplier deleted successfully"})
}

func SupplierRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/Supplier", func(c *fiber.Ctx) error {
		return LookSuppliers(db, c)
	})

	app.Get("/Supplier/:id", func(c *fiber.Ctx) error {
		return FindSupplier(db, c)
	})

	app.Post("/Supplier", func(c *fiber.Ctx) error {
		return AddSupplier(db, c)
	})

	app.Put("/Supplier/:id", func(c *fiber.Ctx) error {
		return UpdateSupplier(db, c)
	})

	app.Delete("/Supplier/:id", func(c *fiber.Ctx) error {
		return DeleteSupplier(db, c)
	})
}
