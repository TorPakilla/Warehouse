package Authentication

import (
	"Api/Models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var JwtKey = []byte("1234")

type Employees struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *fiber.Ctx) error {
	var data Employees
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Bad Request"})
	}

	db := c.Locals("db").(*gorm.DB)

	var employee Models.Employees
	if err := db.Where("username = ?", data.Username).First(&employee).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User not found"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(data.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid Password"})
	}

	expirationTime := time.Now().Add(30 * time.Minute)
	claims := jwt.MapClaims{
		"role":       employee.Role,
		"subject":    employee.Username,
		"issued_at":  time.Now().Unix(),
		"expires_at": expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error"})
	}

	return c.JSON(fiber.Map{"token": tokenString})
}

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	tokenString := authHeader[len("Bearer "):]
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	role := claims["role"].(string)
	if role != "Stock" && role != "Account" && role != "Manager" && role != "Audit" && role != "God" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	}

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	c.Locals("username", claims["subject"])
	c.Locals("role", claims["role"])

	return c.Next()
}
