# 📋 Hướng Dẫn Hệ Thống Logging

## 🎯 Tổng Quan

Hệ thống logging đã được nâng cấp với **dual-logger strategy**:

1. **Console Logger**: Log tất cả lỗi ra console (development)
2. **File Logger**: Chỉ log lỗi nghiêm trọng ra file JSON (production)

---

## 📂 Cấu Trúc File

```
LearnFiber/
├── logger_config.go      # Cấu hình dual-logger system
├── error_handler.go      # Error handler với selective logging
├── logs/
│   ├── errors.log        # File log chính (JSON format)
│   ├── errors.log.gz     # File backup đã nén
│   └── errors.log.*.gz   # Các file backup cũ
```

---

## ⚙️ Cấu Hình Log Rotation

File log tự động rotate theo cấu hình:

| Tham số | Giá trị | Mô tả |
|---------|---------|-------|
| **MaxSize** | 10 MB | Tự động rotate khi file đạt 10MB |
| **MaxBackups** | 5 | Giữ tối đa 5 file backup |
| **MaxAge** | 30 ngày | Xóa file cũ hơn 30 ngày |
| **Compress** | true | Nén file backup thành `.gz` |

---

## 🎨 Phân Loại Lỗi

### ❌ Lỗi Nghiêm Trọng (Log vào File + Console)

| Error Type | Mô tả | HTTP Code |
|------------|-------|-----------|
| **PanicError** | Panic recovered | 500 |
| **SystemError** | Database, filesystem, etc. | 500 |
| **ExternalError** | API/Service bên ngoài lỗi | 502-504 |

### ⚠️ Lỗi Thông Thường (Chỉ log Console)

| Error Type | Mô tả | HTTP Code |
|------------|-------|-----------|
| **ValidationError** | Validation thất bại | 400 |
| **BusinessError** | Business logic | 4xx |
| **AuthError** | Authentication/Authorization | 401-403 |

---

## 📝 Format Log File (JSON)

Mỗi log entry có cấu trúc JSON như sau:

```json
{
  "timestamp": "2025-11-07T10:30:45+07:00",
  "level": "error",
  "message": "Panic recovered: runtime error: index out of range [10] with length 3",
  "error_type": "PANIC",
  "status_code": 500,
  "path": "GET /panic/index",
  "panic_value": {},
  "function": "main.GetElement",
  "file": "main.go:94",
  "call_chain": [
    "main.GetElement (main.go:94)",
    "main.callW (main.go:120)",
    "main.callZ (main.go:115)",
    "main.callY (main.go:110)",
    "main.callX (main.go:105)",
    "main.logrus3Handler (main.go:99)"
  ]
}
```

### Các Field Quan Trọng:

- **timestamp**: Thời gian xảy ra lỗi (RFC3339 format)
- **level**: Mức độ log (error, warn, info)
- **message**: Thông báo lỗi chi tiết
- **error_type**: Loại lỗi (PANIC, SYSTEM, EXTERNAL, BUSINESS, VALIDATION, AUTH)
- **status_code**: HTTP status code
- **path**: HTTP method và endpoint gây lỗi
- **function**: Function gây lỗi
- **file**: File và line number
- **call_chain**: Full call stack (chỉ có khi panic)
- **panic_value**: Giá trị panic (chỉ có khi panic)
- **cause**: Lỗi gốc (nếu có wrapped error)

---

## 🚀 Cách Sử Dụng

### 1. Chạy Server

```bash
go run .
# hoặc
./learnfiber
```

### 2. Test Lỗi Nghiêm Trọng (Sẽ log vào file)

#### Panic Error:
```bash
curl http://localhost:8081/panic/division
curl http://localhost:8081/panic/index
curl http://localhost:8081/panic/stack
```

#### System Error:
```bash
curl http://localhost:8081/error/system
```

#### External Error:
```bash
curl http://localhost:8081/error/external?service=payment
```

### 3. Test Lỗi Thông Thường (Chỉ log console)

#### Validation Error:
```bash
curl http://localhost:8081/error/validation
curl "http://localhost:8081/error/validation?age=abc"
curl "http://localhost:8081/error/validation?age=15"
```

#### Business Error:
```bash
curl "http://localhost:8081/error/business?product_id=123"
```

#### Auth Error:
```bash
curl http://localhost:8081/error/auth
```

---

## 📊 Xem Log File

### Xem log realtime:
```bash
tail -f logs/errors.log
```

### Parse JSON với jq:
```bash
# Xem log đẹp
cat logs/errors.log | jq '.'

# Filter theo error_type
cat logs/errors.log | jq 'select(.error_type == "PANIC")'

# Đếm số lỗi theo type
cat logs/errors.log | jq -r '.error_type' | sort | uniq -c

# Lấy 10 lỗi gần nhất
cat logs/errors.log | jq -s 'sort_by(.timestamp) | reverse | .[0:10]'
```

---

## 🔧 Tuỳ Chỉnh

### Thay đổi cấu hình Log Rotation trong `logger_config.go`:

```go
logFile := &lumberjack.Logger{
    Filename:   "logs/errors.log",
    MaxSize:    50,    // Tăng lên 50MB
    MaxBackups: 10,    // Giữ 10 file backup
    MaxAge:     60,    // Giữ 60 ngày
    Compress:   true,
    LocalTime:  true,
}
```

### Tắt PrettyPrint cho production (tiết kiệm dung lượng):

Trong `logger_config.go`, dòng 53:

```go
fileLogger.SetFormatter(&logrus.JSONFormatter{
    TimestampFormat: time.RFC3339,
    PrettyPrint:     false, // ← Tắt pretty print cho production
    FieldMap: logrus.FieldMap{
        logrus.FieldKeyTime:  "timestamp",
        logrus.FieldKeyLevel: "level",
        logrus.FieldKeyMsg:   "message",
        logrus.FieldKeyFunc:  "function",
    },
})
```

### Thay đổi log level:

```go
// Console logger - log tất cả từ Debug trở lên
consoleLogger.SetLevel(logrus.DebugLevel)

// File logger - chỉ log Error trở lên (production)
fileLogger.SetLevel(logrus.ErrorLevel)
```

### Thêm loại lỗi mới:

1. Thêm ErrorType trong `error_handler.go`:
```go
const (
    BusinessError   ErrorType = "BUSINESS"
    SystemError     ErrorType = "SYSTEM"
    YourNewError    ErrorType = "YOUR_NEW_ERROR" // ← Thêm mới
    // ...
)
```

2. Thêm factory function:
```go
func NewYourNewError(code int, msg string) *AppError {
    file, line, function := getCallerInfo(1)
    return &AppError{
        Type:    YourNewError,
        Code:    code,
        Message: msg,
        Details: map[string]interface{}{
            "function": function,
            "file":     fmt.Sprintf("%s:%d", file, line),
        },
    }
}
```

3. Cập nhật `isSevereError()` nếu cần log vào file:
```go
func isSevereError(errType ErrorType) bool {
    switch errType {
    case PanicError, SystemError, ExternalError, YourNewError: // ← Thêm nếu cần
        return true
    default:
        return false
    }
}
```

---

## 💡 Best Practices

### ✅ Nên:

1. **Log đúng mức độ**:
   - Panic, System, External → File (critical)
   - Business, Validation, Auth → Console only

2. **Structured logging**:
   - Luôn dùng JSON format cho production
   - Thêm metadata đầy đủ (request_id, error_type, location)

3. **Request tracing**:
   - Sử dụng `request_id` để trace lỗi
   - Log cả request path và HTTP method

4. **Log rotation**:
   - Enable auto rotation
   - Compress backup files
   - Set MaxAge để tự động cleanup

5. **Error handling**:
   - Dùng factory functions (`NewBusinessError`, `NewSystemError`, etc.)
   - Wrap errors để giữ nguyên cause
   - Return AppError từ handlers

### ❌ Không nên:

1. **Log quá nhiều**:
   - ❌ Log validation error vào file
   - ❌ Log debug info trong production
   - ❌ Log request/response body mặc định

2. **Bảo mật**:
   - ❌ KHÔNG log password, token, API keys
   - ❌ KHÔNG log sensitive user data
   - ❌ KHÔNG log credit card info

3. **Performance**:
   - ❌ Log synchronously trong critical path
   - ❌ PrettyPrint trong production (tốn disk)
   - ❌ Log quá nhiều fields không cần thiết

4. **Error handling**:
   - ❌ Panic mà không có recovery
   - ❌ Swallow errors (bắt mà không xử lý)
   - ❌ Trả raw error message cho client

---

## 🎯 Ưu Điểm Của Hệ Thống

1. ✅ **Structured logging**: JSON dễ parse và phân tích
2. ✅ **Selective logging**: Chỉ log lỗi nghiêm trọng vào file
3. ✅ **Auto rotation**: Không lo file quá lớn
4. ✅ **Compressed backup**: Tiết kiệm disk space
5. ✅ **Request tracing**: Dễ debug với `request_id`
6. ✅ **Call stack**: Biết chính xác nơi gây panic
7. ✅ **Dual output**: Console (dev) + File (prod)

---

## 📈 Tương Lai - Mở Rộng

Có thể nâng cấp thêm:

### 1. Database Logging
```go
// Log vào PostgreSQL/MySQL để query và phân tích
type ErrorLog struct {
    ID        uint      `gorm:"primaryKey"`
    Timestamp time.Time
    Level     string
    ErrorType string
    Message   string
    Path      string
    RequestID string
}
```

### 2. External Services Integration
- **ELK Stack**: Elasticsearch + Logstash + Kibana
- **Datadog**: Real-time monitoring và alerting
- **Sentry**: Error tracking với source maps
- **CloudWatch**: AWS native logging

### 3. Alert System
```go
// Alert khi có lỗi nghiêm trọng
if appErr.Type == PanicError || appErr.Type == SystemError {
    alerting.SendSlackNotification(appErr)
    alerting.SendEmail(appErr)
}
```

### 4. Metrics & Analytics
- Số lỗi theo error_type
- Top endpoints gây lỗi nhiều nhất
- Response time distribution
- Error rate per minute/hour

### 5. Distributed Tracing
- OpenTelemetry integration
- Trace requests qua nhiều microservices
- Visualize request flow

---

## 🐛 Troubleshooting

### ❌ Không tạo được file log?

**Nguyên nhân**: Thiếu quyền ghi vào thư mục `logs/`

**Giải pháp**:
```bash
mkdir -p logs
chmod 755 logs
```

### ❌ File log quá lớn?

**Nguyên nhân**: MaxSize quá cao hoặc quá nhiều lỗi

**Giải pháp**:
1. Giảm `MaxSize` xuống (vd: 5MB)
2. Giảm `MaxBackups` hoặc `MaxAge`
3. Tắt PrettyPrint (tiết kiệm 30-40% dung lượng)
4. Kiểm tra tại sao có quá nhiều lỗi

### ❌ Không thấy log trong file?

**Nguyên nhân**: Lỗi không thuộc loại nghiêm trọng

**Giải pháp**:
1. Check console output (tất cả lỗi đều log ra console)
2. Xác nhận lỗi là Panic/System/External (chỉ 3 loại này log vào file)
3. Kiểm tra log level: `fileLogger.SetLevel(logrus.ErrorLevel)`

### ❌ Stack trace không đúng?

**Nguyên nhân**: Skip frame không chính xác

**Giải pháp**:
1. Kiểm tra `getCallerInfo(skip)` - skip phải đúng số frame
2. Với panic, dùng `getActualPanicLocation()` thay vì runtime.Caller

### ❌ Log rotation không hoạt động?

**Nguyên nhân**: File đang bị lock hoặc config sai

**Giải pháp**:
```go
logFile := &lumberjack.Logger{
    Filename:   "logs/errors.log",
    MaxSize:    10,     // Đơn vị: MB
    MaxBackups: 5,      // Số lượng file backup
    MaxAge:     30,     // Đơn vị: ngày
    Compress:   true,   // Bắt buộc để nén
    LocalTime:  true,   // Dùng local timezone
}
```

### ❌ Performance chậm khi log?

**Nguyên nhân**: Synchronous I/O blocking

**Giải pháp**:
1. Sử dụng buffered writer
2. Log async với goroutine (cẩn thận với data race)
3. Tắt PrettyPrint
4. Giảm số fields không cần thiết

