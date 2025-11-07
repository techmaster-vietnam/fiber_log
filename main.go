package main

import (
	"fmt"
	"html/template"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// ============================================================================
// Global Variables
// ============================================================================
var (
	homeTemplate *template.Template
)

// init khởi tạo logger và templates
func init() {
	initLoggers() // Khởi tạo dual logger system (console + file)
	initTemplates()
}

// initTemplates khởi tạo HTML templates
func initTemplates() {
	var err error
	homeTemplate, err = template.ParseFiles("templates/home.html")
	if err != nil {
		panic(fmt.Sprintf("Failed to load templates: %v", err))
	}
}

// ============================================================================
// Main
// ============================================================================
func main() {
	app := fiber.New(fiber.Config{
		AppName: "LearnFiber - Logrus Demo",
	})

	// Middleware
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(ErrorHandlerMiddleware())

	// Routes - Home
	app.Get("/", homeHandler)

	// Routes - Panic Errors
	app.Get("/panic/division", logrusHandler)
	app.Get("/panic/index", logrus2Handler)
	app.Get("/panic/stack", logrus3Handler)

	// Routes - Custom Errors
	app.Get("/error/business", businessErrorHandler)
	app.Get("/error/system", systemErrorHandler)
	app.Get("/error/validation", validationErrorHandler)
	app.Post("/error/validation-body", validationBodyHandler)
	app.Get("/error/auth", authErrorHandler)
	app.Get("/error/external", externalErrorHandler)

	// Start server
	fmt.Println("Server starting on http://localhost:8081")
	if err := app.Listen(":8081"); err != nil {
		panic(err)
	}
}

func homeHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return homeTemplate.Execute(c.Response().BodyWriter(), nil)
}

// ============================================================================
// Handlers - Giờ rất gọn, chỉ logic thực tế!
// ============================================================================
func logrusHandler(c *fiber.Ctx) error {
	// Logic thực tế - sẽ gây panic chia cho 0
	denominator := 0
	result := 100 / denominator
	return c.JSON(fiber.Map{"result": result})
}

func logrus2Handler(c *fiber.Ctx) error {
	// Logic thực tế - sẽ gây panic index out of range
	element := GetElement()
	return c.JSON(fiber.Map{"element": element})
}

// GetElement truy cập phần tử mảng không tồn tại
func GetElement() int {
	arr := []int{1, 2, 3}
	return arr[10] // Index out of range - panic tại đây!
}

func logrus3Handler(c *fiber.Ctx) error {
	// Logic thực tế - deep call stack sẽ gây panic
	result := callX()
	return c.JSON(fiber.Map{"result": result})
}

// callX gọi callY
func callX() int {
	return callY()
}

// callY gọi callZ
func callY() int {
	return callZ()
}

// callZ gọi GetElement
func callZ() int {
	return callW() // Panic sẽ xảy ra trong GetElement, không phải ở đây!
}

// callW gọi GetElement
func callW() int {
	return GetElement() // Panic sẽ xảy ra trong GetElement, không phải ở đây!
}

// ============================================================================
// Demo Custom Error Handlers
// ============================================================================

// businessErrorHandler - Demo lỗi business logic (sản phẩm hết hàng)
func businessErrorHandler(c *fiber.Ctx) error {
	productID := c.Query("product_id", "unknown")

	// Giả lập kiểm tra sản phẩm
	if productID == "123" {
		return NewBusinessError(404, fmt.Sprintf("Sản phẩm ID=%s đã hết hàng", productID))
	}

	return c.JSON(fiber.Map{
		"message":    "Sản phẩm có sẵn",
		"product_id": productID,
	})
}

// systemErrorHandler - Demo lỗi hệ thống (database, file system, etc.)
func systemErrorHandler(c *fiber.Ctx) error {
	// Giả lập lỗi database connection
	err := fmt.Errorf("connection refused: database is down")
	return NewSystemError(err)
}

// validationErrorHandler - Demo lỗi validation (query params)
func validationErrorHandler(c *fiber.Ctx) error {
	age := c.Query("age", "")

	if age == "" {
		return NewValidationError("Thiếu tham số 'age'", map[string]interface{}{
			"field":    "age",
			"required": true,
		})
	}

	// Kiểm tra age phải là số
	var ageInt int
	if _, err := fmt.Sscanf(age, "%d", &ageInt); err != nil {
		return NewValidationError("Tham số 'age' phải là số nguyên", map[string]interface{}{
			"field":    "age",
			"type":     "integer",
			"received": age,
		})
	}

	if ageInt < 18 {
		return NewValidationError("Tuổi phải >= 18", map[string]interface{}{
			"field":    "age",
			"min":      18,
			"received": ageInt,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Validation thành công",
		"age":     ageInt,
	})
}

// User struct cho demo validation body
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// validationBodyHandler - Demo lỗi validation (request body)
func validationBodyHandler(c *fiber.Ctx) error {
	var user User

	// Parse body
	if err := c.BodyParser(&user); err != nil {
		return NewValidationError("Request body không hợp lệ", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Validate fields
	if user.Name == "" {
		return NewValidationError("Tên không được để trống", map[string]interface{}{
			"field":    "name",
			"required": true,
		})
	}

	if user.Email == "" {
		return NewValidationError("Email không được để trống", map[string]interface{}{
			"field":    "email",
			"required": true,
		})
	}

	if user.Age < 18 {
		return NewValidationError("Tuổi phải >= 18", map[string]interface{}{
			"field":    "age",
			"min":      18,
			"received": user.Age,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Tạo user thành công",
		"user":    user,
	})
}

// authErrorHandler - Demo lỗi authentication/authorization
func authErrorHandler(c *fiber.Ctx) error {
	token := c.Get("Authorization")

	// Kiểm tra token có tồn tại không
	if token == "" {
		return NewAuthError(401, "Unauthorized: Missing authorization token")
	}

	// Giả lập kiểm tra token không hợp lệ
	if token != "Bearer valid-token-123" {
		return NewAuthError(401, "Unauthorized: Invalid token")
	}

	// Giả lập kiểm tra quyền truy cập
	role := c.Get("X-User-Role")
	if role != "admin" {
		return NewAuthError(403, "Forbidden: Insufficient permissions")
	}

	return c.JSON(fiber.Map{
		"message": "Authentication thành công",
		"role":    role,
	})
}

// externalErrorHandler - Demo lỗi từ external API/service
func externalErrorHandler(c *fiber.Ctx) error {
	// Giả lập gọi external API thất bại
	service := c.Query("service", "payment")

	err := fmt.Errorf("timeout after 30s")

	var statusCode int
	var message string

	switch service {
	case "payment":
		statusCode = 502
		message = "Payment gateway không phản hồi"
	case "shipping":
		statusCode = 503
		message = "Shipping service đang bảo trì"
	case "notification":
		statusCode = 504
		message = "Notification service timeout"
	default:
		statusCode = 502
		message = "External service không khả dụng"
	}

	return NewExternalError(statusCode, message, err)
}
