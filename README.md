# Fiber Log - GoErrorKit Demo

Demo ứng dụng Fiber tích hợp **goerrorkit** để xử lý và logging lỗi chuyên nghiệp.

## 🔧 Tích hợp và cấu hình GoErrorKit vào ứng dụng Fiber

### Bước 1: Cài đặt package

```bash
go get github.com/techmaster-vietnam/goerrorkit
```

**Tác dụng**: Tải về thư viện GoErrorKit vào dự án Go của bạn.

### Bước 2: Import các package cần thiết

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/techmaster-vietnam/goerrorkit"
    "github.com/techmaster-vietnam/goerrorkit/fiberadapter"
)
```

**Tác dụng**: 
- `goerrorkit`: Package chính cung cấp các hàm tạo và xử lý lỗi
- `fiberadapter`: Adapter tích hợp GoErrorKit với Fiber framework

### Bước 3: Khởi tạo Logger

```go
goerrorkit.InitLogger(goerrorkit.LoggerOptions{
    ConsoleOutput: true,           // Log ra console
    FileOutput:    true,            // Log ra file
    FilePath:      "logs/errors.log",
    JSONFormat:    true,            // Format JSON
    MaxFileSize:   10,              // 10MB per file
    MaxBackups:    5,               // Giữ 5 file backup
    MaxAge:        30,              // 30 ngày
    LogLevel:      "info",
})
```

**Tác dụng**: 
- Cấu hình nơi và cách thức ghi log lỗi
- Hỗ trợ log rotation tự động (giới hạn kích thước, số file backup, thời gian lưu trữ)
- Có thể log đồng thời ra console và file

### Bước 4: Cấu hình Stack Trace

```go
goerrorkit.ConfigureForApplication("main")
```

**Tác dụng**: 
- Lọc stack trace chỉ hiển thị code của bạn
- Loại bỏ: Go runtime, Fiber framework, thư viện bên thứ 3
- Kết quả: Stack trace ngắn gọn (5-10 dòng) thay vì 50+ dòng, dễ đọc và debug hơn

### Bước 5: Đăng ký Middleware vào Fiber

```go
app := fiber.New()
app.Use(fiberadapter.ErrorHandler())
```

**Tác dụng**: 
- Tự động bắt mọi error được return từ handlers
- Tự động recover panic và chuyển thành error response
- Tự động log chi tiết error + stack trace vào file đã cấu hình
- Trả về JSON response chuẩn cho client

## 📋 Các Loại Lỗi Được Xử Lý

### 1. **Panic Errors** (Auto-recovered)
- Division by zero
- Index out of range
- Nil pointer dereference
- Deep call stack panics

**Đặc điểm**: Tự động có full call chain, không cần `.WithCallChain()`

### 2. **Business Errors** (`NewBusinessError`)
- Sản phẩm không tồn tại
- Sản phẩm hết hàng
- Không thể hủy đơn đã ship
- Logic nghiệp vụ vi phạm

**Ví dụ**:
```go
goerrorkit.NewBusinessError(404, "Sản phẩm không tồn tại").
    WithData(map[string]interface{}{
        "product_id": productID,
    })
```

### 3. **Validation Errors** (`NewValidationError`)
- Query params không hợp lệ
- Request body sai format
- Số lượng/giá trị ngoài range cho phép
- Thiếu field bắt buộc

**Ví dụ**:
```go
goerrorkit.NewValidationError("Tuổi phải >= 18", map[string]interface{}{
    "field": "age",
    "min": 18,
    "received": 15,
})
```

### 4. **Auth Errors** (`NewAuthError`)
- Missing authorization token
- Invalid token
- Insufficient permissions (403)

**Ví dụ**:
```go
goerrorkit.NewAuthError(401, "Unauthorized: Invalid token")
goerrorkit.NewAuthError(403, "Forbidden: Insufficient permissions")
```

### 5. **External Errors** (`NewExternalError`)
- Payment gateway timeout
- External API không phản hồi
- Third-party service lỗi

**Ví dụ**:
```go
goerrorkit.NewExternalError(504, "Payment gateway timeout", err).
    WithData(map[string]interface{}{
        "order_id": orderID,
        "timeout": "30s",
    })
```

### 6. **System Errors** (`NewSystemError`)
- Database connection failed
- File system errors
- Internal server errors

**Ví dụ**:
```go
goerrorkit.NewSystemError(err).WithData(map[string]interface{}{
    "database": "postgres",
    "host": "localhost:5432",
})
```

## 🎯 Tính Năng Nổi Bật

### WithCallChain()
Thêm full call chain cho non-panic errors:

```go
return goerrorkit.NewValidationError("Dữ liệu không hợp lệ", nil).
    WithCallChain()  // ⭐ Thêm call chain để debug dễ dàng
```

**Kết quả log**:
```json
{
  "location": "main.go:validateOrderData:574",
  "call_chain": [
    "main.go:complexErrorHandler:547",
    "main.go:processOrderData:560",
    "main.go:validateOrderData:574"
  ]
}
```

### WithData()
Thêm context data vào error:

```go
err.WithData(map[string]interface{}{
    "product_id": "123",
    "requested": 10,
    "available": 5,
})
```

## 🚀 Chạy Demo

```bash
go run main.go
```

**Test endpoints**:
- `GET /panic/division` - Panic auto-recovered
- `GET /product/999` - Business error (không tồn tại)
- `GET /product/123/check-stock` - Business error (hết hàng)
- `GET /error/validation?age=15` - Validation error
- `GET /error/auth` - Auth error (missing token)
- `POST /order/ORD-123/payment?amount=20000` - External error (timeout)
- `GET /error/complex` - Complex error với call chain

**Xem logs**: `tail -f logs/errors.log`

## 📂 Cấu Trúc

```
fiber_log/
├── main.go              # Setup + handlers
├── services/
│   ├── product_service.go   # Business logic sản phẩm
│   └── order_service.go     # Business logic đơn hàng
└── logs/
    └── errors.log       # Error logs (JSON format)
```

## 🔍 Log Format

Mỗi error được log với đầy đủ thông tin:

```json
{
  "timestamp": "2025-11-11T10:30:45+07:00",
  "level": "error",
  "error_type": "BusinessError",
  "message": "Sản phẩm đã hết hàng",
  "status_code": 400,
  "location": "services/product_service.go:CheckStock:57",
  "data": {
    "product_id": "123",
    "product_name": "iPhone 15"
  },
  "request_id": "abc123...",
  "http_context": {
    "method": "GET",
    "path": "/product/123/check-stock",
    "ip": "127.0.0.1"
  }
}
```

