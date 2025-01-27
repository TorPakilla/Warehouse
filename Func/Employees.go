package Func

import (
	"Api/Authentication"
	"Api/Models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ฟังก์ชันเพิ่มพนักงานใหม่
func AddEmployees(db *gorm.DB, c *fiber.Ctx) error {
	type UserRequest struct {
		Username string  `json:"username"`
		Password string  `json:"password"`
		Role     string  `json:"role"`
		Name     string  `json:"name"`
		BranchID string  `json:"branch_id"`
		Salary   float64 `json:"salary"`
	}

	var req UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	// ตรวจสอบค่าที่ว่าง
	if strings.TrimSpace(req.Username) == "" ||
		strings.TrimSpace(req.Password) == "" ||
		strings.TrimSpace(req.Role) == "" ||
		strings.TrimSpace(req.Name) == "" ||
		strings.TrimSpace(req.BranchID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All fields are required"})
	}

	branchUUID, err := uuid.Parse(req.BranchID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid BranchID format. Must be a valid UUID"})
	}

	var branch Models.Branches
	if err := db.Table("Branches").Where("branch_id = ?", branchUUID).First(&branch).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "BranchID not found"})
	}

	// Hash Password ก่อนบันทึก
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password: " + err.Error()})
	}

	user := Models.Employees{
		EmployeesID: uuid.New(),
		Username:    req.Username,
		Password:    string(hashedPassword),
		Role:        req.Role,
		Name:        req.Name,
		BranchID:    branchUUID,
		Salary:      req.Salary,
		CreatedAt:   time.Now(),
	}

	if err := db.Table("Employees").Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create employee: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Employee created successfully", "employee": user})
}

// ดึงข้อมูลพนักงาน
func LookEmployees(db *gorm.DB, c *fiber.Ctx) error {
	var employees []Models.Employees

	if err := db.Preload("Branch").Find(&employees).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to find employees: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{"Data": employees})
}

// แก้ไขพนักงาน
func UpdateEmployees(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var user Models.Employees

	if err := db.Table("Employees").Where("employees_id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	type UserRequest struct {
		Username string  `json:"username"`
		Password string  `json:"password"`
		Role     string  `json:"role"`
		Name     string  `json:"name"`
		BranchID string  `json:"branch_id"`
		Salary   float64 `json:"salary"`
	}

	var req UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	// อัปเดตฟิลด์ที่ได้รับ
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password: " + err.Error()})
		}
		user.Password = string(hashedPassword)
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.BranchID != "" {
		branchUUID, err := uuid.Parse(req.BranchID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid BranchID format"})
		}
		user.BranchID = branchUUID
	}
	if req.Salary != 0 {
		user.Salary = req.Salary
	}

	if err := db.Table("Employees").Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update employee: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Employee updated successfully", "user": user})
}

// ลบพนักงาน
func DeleteEmployees(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var user Models.Employees
	if err := db.Where("employees_id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	if err := db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete user: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": "Succeed"})
}

// Routes สำหรับพนักงาน
func EmployeesRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/Employees", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return LookEmployees(db, c)
	})

	app.Post("/Employees", func(c *fiber.Ctx) error {
		return AddEmployees(db, c)
	})

	app.Put("/Employees/:id", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return UpdateEmployees(db, c)
	})

	app.Delete("/Employees/:id", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return DeleteEmployees(db, c)
	})
}
