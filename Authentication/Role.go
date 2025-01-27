package Authentication

import (
	"Api/Models"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ดึงค่าคีย์ JWT จาก Environment Variable
var JwtKey = []byte(os.Getenv("JWT_SECRET"))

// โครงสร้างข้อมูลที่ใช้รับสำหรับ Login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ฟังก์ชันสำหรับ Login
func Login(c *fiber.Ctx) error {
	// รับข้อมูล JSON ที่ผู้ใช้ส่งมา
	var data LoginRequest
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid JSON"})
	}

	// ตรวจสอบว่าข้อมูลครบถ้วน
	if data.Username == "" || data.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Username and Password are required"})
	}

	// ดึง Database จาก Context
	db := c.Locals("db").(*gorm.DB)

	// ค้นหาผู้ใช้ในฐานข้อมูล
	var employee Models.Employees
	if err := db.Where("username = ?", data.Username).First(&employee).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User not found"})
	}

	// ตรวจสอบรหัสผ่าน
	if err := bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(data.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid Password"})
	}

	// สร้าง JWT Token
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

	// ส่ง JWT Token กลับ
	return c.JSON(fiber.Map{"token": tokenString})
}

// Middleware สำหรับตรวจสอบ Token
func AuthMiddleware(c *fiber.Ctx) error {
	// ดึงค่า Authorization Header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	// ตรวจสอบรูปแบบของ Token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid Token Format"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := jwt.MapClaims{}

	// ตรวจสอบความถูกต้องของ Token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	// ตรวจสอบวันหมดอายุของ Token
	exp, ok := claims["exp"].(float64)
	if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Token expired"})
	}

	// ตรวจสอบสิทธิ์ (Role)
	role, ok := claims["role"].(string)
	if !ok || !isValidRole(role) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Permission Denied"})
	}

	// เก็บข้อมูลลงใน Context
	c.Locals("username", claims["username"])
	c.Locals("role", claims["role"])

	// ดำเนินการต่อ
	return c.Next()
}

// ฟังก์ชันตรวจสอบ Role
func isValidRole(role string) bool {
	validRoles := map[string]bool{
		"Stock":   true,
		"Account": true,
		"Manager": true,
		"Audit":   true,
		"God":     true,
	}
	return validRoles[role]
}
