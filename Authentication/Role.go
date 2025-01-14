package Authentication

import (
	"Api/Models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var JwtKey = []byte(os.Getenv("1234"))

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func Login(c *fiber.Ctx) error {
	var data LoginRequest
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Bad Request"})
	}

	if data.Username == "" || data.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Username and Password are required"})
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
		"role":     employee.Role,
		"username": employee.Username,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
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

	if len(authHeader) < len("Bearer ")+1 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid Token Format"})
	}
	tokenString := authHeader[len("Bearer "):]

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	exp, ok := claims["exp"].(float64)
	if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Token expired"})
	}

	role, ok := claims["role"].(string)
	if !ok || !isValidRole(role) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	}

	c.Locals("username", claims["username"])
	c.Locals("role", claims["role"])

	return c.Next()
}

func isValidRole(role string) bool {
	validRoles := []string{"Stock", "Account", "Manager", "Audit", "God"}
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

func LoginHandler(c *fiber.Ctx) error {
	// ดึงข้อมูล username/password จาก body
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// ตรวจสอบ username/password
	if loginRequest.Username == "admin" && loginRequest.Password == "password" {
		token := "mock-jwt-token" // สร้างหรือใช้ JWT จริง
		return c.JSON(fiber.Map{"token": token})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
}
