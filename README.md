# Đọc kỹ hướng dẫn sử dụng trước khi dùng. Bác sỹ hay bảo vậy.

Vấn đề của chúng ta hiện này là gì:
- Thường dùng hàm log thông thường để báo lỗi ra console. Khi lên production, không thể xem lại lịch sử lỗi
- Log lỗi chung chung không biết dòng nào gây lỗi, danh sách các hàm gọi lồng nhau cũng không biết nốt
- Không phân loại được lỗi. Lỗi validation khác lỗi hệ thống và lỗi panic đúng không?
- Không cung cấp đủ thông tin về lỗi kiểu như giá trị biến tại thời điểm lỗi

Tóm lại chúng ta code báo lỗi chỉ 

Dự án demo về Error Handling và Logging System với Fiber framework.

## Mô tả

Ứng dụng web này demo cách xử lý lỗi chuyên nghiệp trong Go với:

1. **Custom Error Types** - Phân loại lỗi rõ ràng (Panic, System, External, Business, Validation, Auth)
2. **Error Handler Middleware** - Xử lý lỗi tập trung với panic recovery
3. **Dual Logger Strategy** - Console (development) + File (production)
4. **Selective Logging** - Chỉ log lỗi nghiêm trọng vào file
5. **Stack Trace Analysis** - Tự động phân tích call stack khi panic

## Tính năng

Hệ thống xử lý và logging cung cấp:

- ✅ **Panic Recovery**: Tự động bắt và xử lý panic
- ✅ **Call Stack Tracking**: Trace đầy đủ call chain khi xảy ra panic
- ✅ **Structured Logging**: JSON format với đầy đủ metadata
- ✅ **Log Rotation**: Tự động rotate và nén file log
- ✅ **Error Classification**: Phân loại lỗi theo mức độ nghiêm trọng
- ✅ **Request Tracing**: Track error với request_id
- ✅ **Location Detection**: Xác định chính xác nơi gây lỗi (file:line)

## Cài đặt

### Yêu cầu

- Go 1.21 trở lên

### Các bước cài đặt

1. Clone repository hoặc cd vào thư mục dự án:

```bash
cd /Users/cuong/CODE/LearnFiber
```

2. Cài đặt dependencies:

```bash
go mod download
```

3. Build ứng dụng:

```bash
go build -o learnfiber
```

## Sử dụng

### Chạy server

```bash
go run .
```

Hoặc chạy file đã build:

```bash
./learnfiber
```

Server sẽ khởi động tại: **http://localhost:8081**

### Các Endpoints

#### 🏠 Trang chủ
- `GET /` - Trang chủ với UI đẹp, danh sách đầy đủ các endpoints

#### ⚡ Panic Errors (Lỗi nghiêm trọng - log vào file)
| Endpoint | Mô tả | HTTP Code |
|----------|-------|-----------|
| `GET /panic/division` | Division by zero panic | 500 |
| `GET /panic/index` | Index out of range panic | 500 |
| `GET /panic/stack` | Deep call stack panic (X→Y→Z→W→GetElement) | 500 |

#### 💼 Business Errors (Lỗi logic nghiệp vụ)
| Endpoint | Mô tả | HTTP Code |
|----------|-------|-----------|
| `GET /error/business?product_id=123` | Sản phẩm hết hàng | 404 |

#### ✅ Validation Errors (Lỗi validation)
| Endpoint | Mô tả | HTTP Code |
|----------|-------|-----------|
| `GET /error/validation` | Thiếu hoặc sai query params | 400 |
| `POST /error/validation-body` | Validation request body | 400 |

#### 🔐 Auth Errors (Lỗi xác thực)
| Endpoint | Mô tả | HTTP Code |
|----------|-------|-----------|
| `GET /error/auth` | Missing/invalid token hoặc insufficient permissions | 401-403 |

#### ⚙️ System Errors (Lỗi hệ thống - log vào file)
| Endpoint | Mô tả | HTTP Code |
|----------|-------|-----------|
| `GET /error/system` | Database/filesystem error | 500 |

#### 🌐 External Errors (Lỗi external service - log vào file)
| Endpoint | Mô tả | HTTP Code |
|----------|-------|-----------|
| `GET /error/external?service=payment` | Payment gateway error | 502 |
| `GET /error/external?service=shipping` | Shipping service unavailable | 503 |
| `GET /error/external?service=notification` | Notification timeout | 504 |

### Ví dụ sử dụng

```bash
# 1. Mở trang chủ trong browser
open http://localhost:8081/

# 2. Test Panic Errors
curl http://localhost:8081/panic/division
curl http://localhost:8081/panic/index
curl http://localhost:8081/panic/stack

# 3. Test Business Errors
curl http://localhost:8081/error/business?product_id=123

# 4. Test Validation Errors
curl http://localhost:8081/error/validation
curl "http://localhost:8081/error/validation?age=abc"
curl "http://localhost:8081/error/validation?age=15"
curl "http://localhost:8081/error/validation?age=25"

# 5. Test Validation Body
curl -X POST http://localhost:8081/error/validation-body \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@test.com","age":25}'

# 6. Test Auth Errors
curl http://localhost:8081/error/auth
curl -H "Authorization: Bearer valid-token-123" \
     -H "X-User-Role: admin" \
     http://localhost:8081/error/auth

# 7. Test System Errors
curl http://localhost:8081/error/system

# 8. Test External Errors
curl http://localhost:8081/error/external?service=payment
curl http://localhost:8081/error/external?service=shipping
curl http://localhost:8081/error/external?service=notification
```

### Xem Log Output

Kiểm tra console để xem log chi tiết:
- **Console**: Tất cả lỗi được log ra console với màu sắc
- **File**: Chỉ lỗi nghiêm trọng (Panic, System, External) được log vào `logs/errors.log`

```bash
# Xem log file realtime
tail -f logs/errors.log

# Parse JSON log với jq
cat logs/errors.log | jq '.'
```

## Kiến Trúc

### Phân loại lỗi (Error Types)

| Error Type | Mã HTTP | Mức độ | Log vào File? |
|------------|---------|---------|---------------|
| **PanicError** | 500 | Critical | ✅ Có |
| **SystemError** | 500 | Critical | ✅ Có |
| **ExternalError** | 502-504 | Critical | ✅ Có |
| **BusinessError** | 4xx | Warning | ❌ Không |
| **ValidationError** | 400 | Warning | ❌ Không |
| **AuthError** | 401-403 | Info | ❌ Không |

### Luồng xử lý lỗi

1. **Request** → Fiber Router → Handler
2. **Handler** throws error hoặc panic
3. **ErrorHandlerMiddleware** bắt error/panic
4. **Classification**: Xác định loại error
5. **Logging**: 
   - Console: Log tất cả
   - File: Chỉ log critical errors
6. **Response**: Trả JSON error cho client

### Dual Logger Strategy

```
┌─────────────────────────────────────┐
│   ErrorHandlerMiddleware            │
│                                     │
│   ┌──────────────────────────┐     │
│   │  Console Logger          │     │
│   │  - Tất cả lỗi           │     │
│   │  - Màu sắc, dễ đọc      │     │
│   │  - Development mode      │     │
│   └──────────────────────────┘     │
│                                     │
│   ┌──────────────────────────┐     │
│   │  File Logger             │     │
│   │  - Chỉ lỗi nghiêm trọng │     │
│   │  - JSON format           │     │
│   │  - Auto rotation         │     │
│   │  - Production mode       │     │
│   └──────────────────────────┘     │
└─────────────────────────────────────┘
```

## Cấu trúc dự án

```
LearnFiber/
├── main.go              # Entry point, routes, handlers
├── error_handler.go     # Custom error types, middleware, log handlers
├── logger_config.go     # Dual logger configuration
├── call_stack_log.go    # Stack trace analysis utilities
├── templates/
│   └── home.html        # Beautiful UI homepage
├── logs/
│   ├── errors.log       # JSON log file (auto-rotated)
│   └── errors.log.*.gz  # Compressed backups
├── go.mod               # Module definition
├── go.sum               # Dependencies checksums
├── learnfiber           # Compiled binary
├── README.md            # Documentation (this file)
└── LOGGING_GUIDE.md     # Detailed logging guide
```

## Dependencies

```go
github.com/gofiber/fiber/v2 v2.52.9       // Web framework
github.com/sirupsen/logrus v1.9.3         // Structured logger
gopkg.in/natefinch/lumberjack.v2 v2.2.1   // Log rotation
```

## Công Nghệ Sử Dụng

- **Fiber v2**: Fast HTTP framework, Express-style API
- **Logrus**: Structured logger với JSON formatter
- **Lumberjack**: Log rotation và compression
- **Runtime/Debug**: Stack trace analysis
- **HTML Templates**: Server-side rendering

## Tính Năng Nổi Bật

### 1. Panic Recovery với Call Stack Tracking
Khi xảy ra panic, hệ thống tự động:
- Bắt panic và recover
- Phân tích stack trace
- Xác định chính xác dòng code gây lỗi
- Log đầy đủ call chain
- Trả response thân thiện cho client

### 2. Selective Logging
- **Console**: Log tất cả lỗi cho development
- **File**: Chỉ log lỗi nghiêm trọng (Panic, System, External)
- Tiết kiệm disk space và dễ monitoring

### 3. Log Rotation
- Auto rotate khi file đạt 10MB
- Giữ tối đa 5 backups
- Compress backups thành .gz
- Xóa file cũ hơn 30 ngày

### 4. Request Tracing
Mỗi request có `request_id` unique để trace:
```json
{
  "request_id": "36b9d7d9-9752-4831-aee0-01eee86a41f3",
  "request_path": "GET /panic/index",
  "error_type": "PANIC",
  "message": "Panic recovered: runtime error: index out of range"
}
```

## License

MIT License - Dự án học tập và demo

