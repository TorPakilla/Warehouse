package Func

import (
	"Api/Models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddEmloyees(db *gorm.DB, c *fiber.Ctx) error {
	type UserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `gorm:"check:role_check" json:"role"`
	}

	var req UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	user := Models.Emloyees{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}
	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"New": user})
}

func LookEmloyees(db *gorm.DB, c *fiber.Ctx) error {
	var users []Models.Emloyees
	if err := db.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find users: " + err.Error()})
	}
	return c.JSON(fiber.Map{"This": "User", "Data": users})
}

func FindEmloyees(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var user Models.Emloyees
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(fiber.Map{"This": "User", "Data": user})
}

func DeleteEmloyees(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var user Models.Emloyees
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	if err := db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete user: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Deleted": "Succeed"})
}

func UpdateEmloyees(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var user Models.Emloyees
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	type UserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `gorm:"check:role_check" json:"role"`
	}

	var req UserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format: " + err.Error()})
	}

	user.Username = req.Username
	user.Password = req.Password
	user.Role = req.Role

	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"Updated": "Succeed"})
}

func EmloyeesRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/Emloyees", func(c *fiber.Ctx) error {
		return LookEmloyees(db, c)
	})
	app.Get("/Emloyees", func(c *fiber.Ctx) error {
		return FindEmloyees(db, c)
	})
	app.Post("/Emloyees", func(c *fiber.Ctx) error {
		return AddEmloyees(db, c)
	})
	app.Put("/Emloyees", func(c *fiber.Ctx) error {
		return UpdateEmloyees(db, c)
	})
	app.Delete("/Emloyees", func(c *fiber.Ctx) error {
		return DeleteEmloyees(db, c)
	})
}
