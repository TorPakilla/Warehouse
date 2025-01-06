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
		Username  string  `json:"username"`
		Password  string  `json:"password"`
		Role      string  `json:"role"`
		Name      string  `json:"name"`
		BrancheID string  `json:"brancheid"`
		Salary    float64 `json:"salary"`
	}

	var req UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	CheckRoles := map[string]bool{
		"Admin":   true,
		"User":    true,
		"Manager": true,
		"God":     true,
	}

	if !CheckRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Allowed "})
	}

	user := Models.Employees{
		Username:  req.Username,
		Password:  string(hashedPassword),
		Role:      req.Role,
		Name:      req.Name,
		BrancheID: req.BrancheID,
		Salary:    req.Salary,
	}
	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New": user})
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
	if err := db.Where("employees_id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	type UserRequest struct {
		Username  string  `json:"username"`
		Password  string  `json:"password"`
		Role      string  `json:"role"`
		Name      string  `json:"name"`
		BrancheID string  `json:"brancheid"`
		Salary    float64 `json:"salary"`
	}

	var req UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	CheckRoles := map[string]bool{
		"Stock":   true,
		"Account": true,
		"Manager": true,
		"Audit":   true,
		"God":     true,
	}

	if !CheckRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role. Allowed "})
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
		}
		req.Password = string(hashedPassword)
	}

	user.Username = req.Username
	user.Password = req.Password
	user.Role = req.Role
	user.Name = req.Name
	user.BrancheID = req.BrancheID
	user.Salary = req.Salary

	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": "Succeed"})
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

	app.Post("/Employees", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return AddEmployees(db, c)
	})

	app.Put("/Employees/:id", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return UpdateEmployees(db, c)
	})

	app.Delete("/Employees/:id", Authentication.AuthMiddleware, func(c *fiber.Ctx) error {
		return DeleteEmployees(db, c)
	})
}
