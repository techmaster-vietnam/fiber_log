package main

import (
	"fmt"
	"html/template"
	"strconv"

	"fiber_log/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/techmaster-vietnam/goerrorkit"
	fiberadapter "github.com/techmaster-vietnam/goerrorkit/adapters/fiber"
)

// ============================================================================
// Global Variables
// ============================================================================
var (
	homeTemplate   *template.Template
	productService *services.ProductService
	orderService   *services.OrderService
)

// init khởi tạo logger và templates
func init() {
	// 1. Initialize logger với custom options
	goerrorkit.InitLogger(goerrorkit.LoggerOptions{
		ConsoleOutput: true,
		FileOutput:    true,
		FilePath:      "logs/errors.log",
		JSONFormat:    true,
		MaxFileSize:   10, // MB
		MaxBackups:    5,
		MaxAge:        30, // days
		LogLevel:      "info",
	})

	// 2. Configure stack trace for this application
	// 🎯 MỤC ĐÍCH: Lọc stack trace để CHỈ HIỂN THỊ code của BẠN, bỏ qua:
	//    - Go runtime code (runtime.*, runtime/debug.*)
	//    - Thư viện bên thứ 3 (fiber, goerrorkit, etc.)
	//
	// ✅ CÁCH DÙNG:
	//    - App đơn giản (1 file main.go):
	//      goerrorkit.ConfigureForApplication("main")
	//
	//    - App với nhiều package (services/, handlers/, models/...):
	//      goerrorkit.ConfigureForApplication("fiber_log")
	//      hoặc cấu hình nhiều packages:
	//      goerrorkit.Configure().IncludePackages("main", "fiber_log/services").Apply()
	//
	// 📊 KẾT QUẢ:
	//    KHÔNG cấu hình: Stack trace dài 50+ dòng (runtime, fiber, goerrorkit...)
	//    CÓ cấu hình:    Stack trace ngắn gọn, chỉ 5-10 dòng CODE CỦA BẠN!
	//
	goerrorkit.ConfigureForApplication("main")

	// 🔧 FLUENT API: Nếu cần thêm các patterns tùy chỉnh, có thể dùng:
	//
	// Cách 1: Shorthand - Nhanh chóng thêm skip patterns
	// goerrorkit.AddSkipPatterns(".RequestID.func", ".Logger.func", "telemetry")
	//
	// Cách 2: Fluent API - Configuration chi tiết hơn
	// goerrorkit.ConfigureForApplication().
	//     SkipPattern(".CustomMiddleware.func").
	//     SkipPackage("internal/metrics").
	//     SkipFunctions("helper", "wrapper").
	//     ShowFullPath(false).
	//     Apply()

	initTemplates()
	initServices()
}

// initServices khởi tạo business services
func initServices() {
	productService = services.NewProductService()
	orderService = services.NewOrderService(productService)
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
		AppName: "FiberLog - GoErrorKit Demo",
	})

	// Middleware
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(fiberadapter.ErrorHandler()) // Sử dụng goerrorkit middleware

	// Routes - Home
	app.Get("/", homeHandler)

	// Routes - Panic Errors
	app.Get("/panic/division", panicDivisionHandler)
	app.Get("/panic/index", panicIndexHandler)
	app.Get("/panic/stack", panicStackHandler)

	// Routes - Custom Errors
	app.Get("/error/business", businessErrorHandler)
	app.Get("/error/system", systemErrorHandler)
	app.Get("/error/validation", validationErrorHandler)
	app.Post("/error/validation-body", validationBodyHandler)
	app.Get("/error/auth", authErrorHandler)
	app.Get("/error/external", externalErrorHandler)
	app.Get("/error/complex", complexErrorWithCallChainHandler)

	// Routes - Service Layer Errors (Demo lỗi từ package khác)
	app.Get("/product/:id", getProductHandler)
	app.Get("/product/:id/check-stock", checkStockHandler)
	app.Post("/product/:id/reserve", reserveProductHandler)
	app.Get("/product/:id/discount", calculateDiscountHandler)
	app.Post("/order/create", createOrderHandler)
	app.Delete("/order/:id/cancel", cancelOrderHandler)
	app.Post("/order/:id/payment", processPaymentHandler)

	// Start server
	fmt.Println("🚀 Server starting on http://localhost:8081")
	fmt.Println("\n📝 Try these endpoints:")
	fmt.Println("  GET  /                                    - Home page")
	fmt.Println("\n  🔥 Panic Demos (auto-recovered):")
	fmt.Println("  GET  /panic/division                      - Division by zero")
	fmt.Println("  GET  /panic/index                         - Index out of range")
	fmt.Println("  GET  /panic/stack                         - Deep call stack panic")
	fmt.Println("\n  ⚠️  Custom Error Demos:")
	fmt.Println("  GET  /error/business?product_id=123       - Business error (hết hàng)")
	fmt.Println("  GET  /error/system                        - System error (database)")
	fmt.Println("  GET  /error/validation?age=15             - Validation error")
	fmt.Println("  POST /error/validation-body               - Body validation")
	fmt.Println("  GET  /error/auth                          - Auth error (token)")
	fmt.Println("  GET  /error/external?service=payment      - External API error")
	fmt.Println("  GET  /error/complex                       - Complex error WITH call_chain ⭐")
	fmt.Println("\n  🛍️  Service Layer Demos:")
	fmt.Println("  GET  /product/999                         - Product not found")
	fmt.Println("  GET  /product/123/check-stock             - Stock check (hết hàng)")
	fmt.Println("  POST /product/456/reserve?quantity=10     - Reserve product")
	fmt.Println("  GET  /product/456/discount?percent=150    - Calculate discount")
	fmt.Println("  POST /order/create?product_id=123&quantity=1  - Create order")
	fmt.Println("  DELETE /order/ORD-shipped/cancel          - Cancel order")
	fmt.Println("  POST /order/ORD-123/payment?amount=20000  - Process payment")
	fmt.Println("\n📄 Check logs/errors.log for detailed error logs")

	if err := app.Listen(":8081"); err != nil {
		panic(err)
	}
}

func homeHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return homeTemplate.Execute(c.Response().BodyWriter(), nil)
}

// ============================================================================
// Panic Handlers - Demonstrate automatic panic recovery
// ============================================================================

func panicDivisionHandler(c *fiber.Ctx) error {
	// This will panic with "integer divide by zero"
	denominator := 0
	result := 100 / denominator // ← Panic location will be captured HERE!
	return c.JSON(fiber.Map{"result": result})
}

func panicIndexHandler(c *fiber.Ctx) error {
	// This will panic with "index out of range"
	element := GetElement() // Panic happens inside GetElement()
	return c.JSON(fiber.Map{"element": element})
}

func GetElement() int {
	arr := []int{1, 2, 3}
	return arr[10] // ← Panic location will be captured HERE!
}

func panicStackHandler(c *fiber.Ctx) error {
	// Deep call stack demo
	result := callX()
	return c.JSON(fiber.Map{"result": result})
}

func callX() int {
	return callY()
}

func callY() int {
	return callZ()
}

func callZ() int {
	return callW()
}

func callW() int {
	return GetElement() // Panic happens here, full call chain will be logged
}

// ============================================================================
// Demo Custom Error Handlers
// ============================================================================

// businessErrorHandler - Demo lỗi business logic (sản phẩm hết hàng)
// Error được throw từ SERVICE LAYER - test GoErrorKit báo đúng vị trí
func businessErrorHandler(c *fiber.Ctx) error {
	productID := c.Query("product_id", "123") // Default 123 để test hết hàng

	// Gọi service - error sẽ được throw từ services/product_service.go
	err := productService.CheckStock(productID)
	if err != nil {
		return err // Propagate error từ service layer
	}

	return c.JSON(fiber.Map{
		"message":    "Sản phẩm còn hàng",
		"product_id": productID,
	})
}

// systemErrorHandler - Demo lỗi hệ thống (database, file system, etc.)
func systemErrorHandler(c *fiber.Ctx) error {
	// Giả lập lỗi database connection
	err := fmt.Errorf("connection refused: database is down")
	return goerrorkit.NewSystemError(err).WithData(map[string]interface{}{
		"database": "postgres",
		"host":     "localhost:5432",
	})
}

// validationErrorHandler - Demo lỗi validation (query params)
func validationErrorHandler(c *fiber.Ctx) error {
	age := c.Query("age", "")

	if age == "" {
		return goerrorkit.NewValidationError("Thiếu tham số 'age'", map[string]interface{}{
			"field":    "age",
			"required": true,
		})
	}

	// Kiểm tra age phải là số
	var ageInt int
	if _, err := fmt.Sscanf(age, "%d", &ageInt); err != nil {
		return goerrorkit.NewValidationError("Tham số 'age' phải là số nguyên", map[string]interface{}{
			"field":    "age",
			"type":     "integer",
			"received": age,
		})
	}

	if ageInt < 18 {
		return goerrorkit.NewValidationError("Tuổi phải >= 18", map[string]interface{}{
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
		return goerrorkit.NewValidationError("Request body không hợp lệ", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Validate fields
	if user.Name == "" {
		return goerrorkit.NewValidationError("Tên không được để trống", map[string]interface{}{
			"field":    "name",
			"required": true,
		})
	}

	if user.Email == "" {
		return goerrorkit.NewValidationError("Email không được để trống", map[string]interface{}{
			"field":    "email",
			"required": true,
		})
	}

	if user.Age < 18 {
		return goerrorkit.NewValidationError("Tuổi phải >= 18", map[string]interface{}{
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
		return goerrorkit.NewAuthError(401, "Unauthorized: Missing authorization token")
	}

	// Giả lập kiểm tra token không hợp lệ
	if token != "Bearer valid-token-123" {
		return goerrorkit.NewAuthError(401, "Unauthorized: Invalid token").WithData(map[string]interface{}{
			"token_length": len(token),
		})
	}

	// Giả lập kiểm tra quyền truy cập
	role := c.Get("X-User-Role")
	if role != "admin" {
		return goerrorkit.NewAuthError(403, "Forbidden: Insufficient permissions").WithData(map[string]interface{}{
			"required_role": "admin",
			"user_role":     role,
		})
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

	return goerrorkit.NewExternalError(statusCode, message, err).WithData(map[string]interface{}{
		"service": service,
		"timeout": "30s",
	})
}

// ============================================================================
// Service Layer Handlers - Demo lỗi từ package khác
// ============================================================================

// getProductHandler - Lấy thông tin sản phẩm
// Test: GET /product/999 (không tồn tại) -> BusinessError từ services/product_service.go
func getProductHandler(c *fiber.Ctx) error {
	productID := c.Params("id")

	// Error sẽ được throw từ ProductService.GetProduct
	product, err := productService.GetProduct(productID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"product": product,
	})
}

// checkStockHandler - Kiểm tra tồn kho
// Test: GET /product/123/check-stock (hết hàng) -> BusinessError từ CheckStock
func checkStockHandler(c *fiber.Ctx) error {
	productID := c.Params("id")

	// Error sẽ được throw từ ProductService.CheckStock
	err := productService.CheckStock(productID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Sản phẩm còn hàng",
	})
}

// reserveProductHandler - Đặt trước sản phẩm
// Test: POST /product/456/reserve?quantity=10 -> ValidationError (không đủ hàng)
func reserveProductHandler(c *fiber.Ctx) error {
	productID := c.Params("id")
	quantityStr := c.Query("quantity", "1")
	quantity, _ := strconv.Atoi(quantityStr)

	// Error sẽ được throw từ ProductService.ReserveProduct
	err := productService.ReserveProduct(productID, quantity)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message":  "Đặt hàng thành công",
		"quantity": quantity,
	})
}

// calculateDiscountHandler - Tính giá sau giảm giá
// Test: GET /product/456/discount?percent=150 -> ValidationError (percent không hợp lệ)
func calculateDiscountHandler(c *fiber.Ctx) error {
	productID := c.Params("id")
	percentStr := c.Query("percent", "10")
	percent, _ := strconv.ParseFloat(percentStr, 64)

	// Error sẽ được throw từ ProductService.CalculateDiscount
	finalPrice, err := productService.CalculateDiscount(productID, percent)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"original_price": "check product",
		"discount":       percent,
		"final_price":    finalPrice,
	})
}

// createOrderHandler - Tạo đơn hàng mới
// Test: POST /order/create?product_id=123&quantity=1 -> BusinessError (hết hàng)
// Test: POST /order/create?product_id=456&quantity=0 -> ValidationError (quantity <= 0)
func createOrderHandler(c *fiber.Ctx) error {
	productID := c.Query("product_id")
	userID := c.Query("user_id", "USER001")
	quantityStr := c.Query("quantity", "1")
	quantity, _ := strconv.Atoi(quantityStr)

	// Error có thể được throw từ nhiều nơi trong OrderService
	order, err := orderService.CreateOrder(productID, userID, quantity)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Đơn hàng đã được tạo",
		"order":   order,
	})
}

// cancelOrderHandler - Hủy đơn hàng
// Test: DELETE /order/ORD-shipped/cancel -> BusinessError (đã ship)
func cancelOrderHandler(c *fiber.Ctx) error {
	orderID := c.Params("id")

	// Error sẽ được throw từ OrderService.CancelOrder
	err := orderService.CancelOrder(orderID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message":  "Đơn hàng đã được hủy",
		"order_id": orderID,
	})
}

// processPaymentHandler - Xử lý thanh toán
// Test: POST /order/ORD-invalid-card/payment?amount=100 -> ExternalError (payment gateway)
// Test: POST /order/ORD-123/payment?amount=20000 -> ExternalError (timeout)
func processPaymentHandler(c *fiber.Ctx) error {
	orderID := c.Params("id")
	amountStr := c.Query("amount", "0")
	amount, _ := strconv.ParseFloat(amountStr, 64)

	// Error có thể được throw từ deep trong call stack (OrderService -> callPaymentGateway)
	err := orderService.ProcessPayment(orderID, amount)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message":  "Thanh toán thành công",
		"order_id": orderID,
		"amount":   amount,
	})
}

// ============================================================================
// Complex Error Handler - Demo WithCallChain()
// ============================================================================

// complexErrorWithCallChainHandler demonstrates using .WithCallChain()
// to add full call chain to non-panic errors for better debugging
//
// 🎯 TÍNH NĂNG: .WithCallChain()
// - Panic errors: Tự động có full call chain (không cần .WithCallChain())
// - Normal errors: MẶC ĐỊNH chỉ có location nơi error được tạo
// - .WithCallChain(): Thêm FULL CALL CHAIN vào normal errors!
//
// 📊 SO SÁNH:
// KHÔNG dùng .WithCallChain():
//
//	location: "fiber_log/main.go:validateOrderData:520"
//
// CÓ dùng .WithCallChain():
//
//	location: "fiber_log/main.go:validateOrderData:520"
//	call_chain: [
//	  "fiber_log/main.go:complexErrorWithCallChainHandler:490",
//	  "fiber_log/main.go:processOrderData:500",
//	  "fiber_log/main.go:validateOrderData:520"
//	]
//
// Test: GET /error/complex
func complexErrorWithCallChainHandler(c *fiber.Ctx) error {
	// Simulate a complex operation with multiple function calls
	result, err := processOrderData()
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Đơn hàng đã được xử lý",
		"result":  result,
	})
}

func processOrderData() (string, error) {
	// Call validation
	if err := validateOrderData(); err != nil {
		return "", err
	}

	// Call inventory check
	if err := checkInventoryData(); err != nil {
		return "", err
	}

	return "success", nil
}

func validateOrderData() error {
	// Simulate validation
	isValid := false

	if !isValid {
		// ⭐ Sử dụng .WithCallChain() để thêm full call chain
		// Giúp trace được: complexErrorWithCallChainHandler → processOrderData → validateOrderData
		return goerrorkit.NewValidationError("Dữ liệu đơn hàng không hợp lệ", map[string]interface{}{
			"reason": "invalid_order_data",
		}).WithCallChain() // ⭐ Thêm call_chain vào error!
	}

	return nil
}

func checkInventoryData() error {
	// Simulate inventory check
	stockAvailable := 0

	if stockAvailable == 0 {
		// ⭐ Chain nhiều methods: WithData() + WithCallChain()
		return goerrorkit.NewBusinessError(422, "Không đủ hàng trong kho").
			WithData(map[string]interface{}{
				"product_id": "PROD-123",
				"requested":  10,
				"available":  0,
				"warehouse":  "WH-01",
			}).
			WithCallChain() // ⭐ Thêm call_chain để trace flow
	}

	return nil
}
