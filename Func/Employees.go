package Func

import (
	"Api/Authentication"
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func AddEmployees(db *gorm.DB, c *fiber.Ctx) error {
	type UserRequest struct {
		Username string  `json:"username"`
		Password string  `json:"password"`
		Role     string  `json:"role"`
		Name     string  `json:"name"`
		BranchID string  `json:"branchid"`
		Salary   float64 `json:"salary"`
	}

	var req UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	if req.Username == "" || req.Password == "" || req.Role == "" || req.Name == "" || req.BranchID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All fields are required"})
	}

	var branch Models.Branches
	if err := db.Where("branch_id = ?", req.BranchID).First(&branch).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "BrancheID not found"})
	}

	validRoles := map[string]bool{
		"Stock":   true,
		"Account": true,
		"Manager": true,
		"Audit":   true,
		"God":     true,
	}

	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Allowed roles are Admin, User, Manager, God."})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	user := Models.Employees{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
		Name:     req.Name,
		BranchID: req.BranchID,
		Salary:   req.Salary,
	}

	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"new_user": user})
}

func LookEmployees(db *gorm.DB, c *fiber.Ctx) error {
	var users []Models.Employees
	if err := db.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find users: " + err.Error()})
	}
	return c.JSON(fiber.Map{"This": "User", "Data": users})
}

func FindEmployees(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var user Models.Employees
	if err := db.Where("employees_id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(fiber.Map{"This": "User", "Data": user})
}

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

func UpdateEmployees(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var user Models.Employees

	// ตรวจสอบว่าผู้ใช้มีอยู่หรือไม่
	if err := db.Where("employees_id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	type UserRequest struct {
		Username string  `json:"username"`
		Password string  `json:"password"`
		Role     string  `json:"role"`
		Name     string  `json:"name"`
		BranchID string  `json:"branchid"`
		Salary   float64 `json:"salary"`
	}

	var req UserRequest
	// แปลง JSON เป็น struct
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	// ตรวจสอบ role ที่อนุญาต
	validRoles := map[string]bool{
		"Stock":   true,
		"Account": true,
		"Manager": true,
		"Audit":   true,
		"God":     true,
	}

	if req.Role != "" && !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Allowed roles are Stock, Account, Manager, Audit, God."})
	}

	// ตรวจสอบว่า BrancheID มีอยู่ในฐานข้อมูลหรือไม่ (ถ้ามีการส่งมา)
	if req.BranchID != "" {
		var branch Models.Branches
		if err := db.Where("branche_id = ?", req.BranchID).First(&branch).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "BrancheID not found"})
		}
	}

	// แฮชรหัสผ่านหากมีการอัปเดต
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
		}
		req.Password = string(hashedPassword)
	}

	// อัปเดตฟิลด์ที่ได้รับ
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Password != "" {
		user.Password = req.Password
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.BranchID != "" {
		user.BranchID = req.BranchID
	}
	if req.Salary != 0 {
		user.Salary = req.Salary
	}

	// บันทึกข้อมูล
	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Update Succeed", "user": user})
}

func EmployeesRoutes(app *fiber.App, db *gorm.DB) {
	app.Use(func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "God" && role != "Manager" {
			return c.Next()
		}

		if role != "Stock" && role != "Account" && role != "Audit" {
			if c.Method() != "GET" && c.Method() != "UPDATE" {
				return c.Next()
			} else {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	})

	app.Get("/Employees", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return LookEmployees(db, c)
	})

	app.Get("/Employees/:id", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return FindEmployees(db, c)
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
