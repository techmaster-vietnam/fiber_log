# Đọc kỹ hướng dẫn sử dụng trước khi dùng. Bác sỹ hay bảo vậy.

Vấn đề của chúng ta hiện này là gì:
- Thường dùng hàm log thông thường để báo lỗi ra console. Khi lên production, không thể xem lại lịch sử lỗi
- Log lỗi chung chung không biết dòng nào gây lỗi, danh sách các hàm gọi lồng nhau cũng không biết nốt
- Không phân loại được lỗi. Lỗi validation khác lỗi hệ thống và lỗi panic đúng không?
- Không cung cấp đủ thông tin về lỗi kiểu như giá trị biến tại thời điểm lỗi

Tóm lại chúng ta code báo lỗi chỉ debug khi phát triển chứ không thực sự báo lỗi thành hệ thống để khi triển khai môi trường production có thể tìm ra nguyên nhân lỗi nhanh để sửa dứt điểm.

Sau một ngày hì hục ra lệnh cho AI tôi đã xong thư viện log lỗi sử dụng lại  hai thư viện chính là "sirupsen/logrus" log ra json và "lumberjack.v2" để nén file log

## Tính năng
1. **Custom Error Types** - Phân loại lỗi rõ ràng (Panic, System, External, Business, Validation, Auth)
2. **Error Handler Middleware** - Xử lý lỗi tập trung với panic recovery
3. **Dual Logger Strategy** - Console (development) + File (production)
4. **Selective Logging** - Chỉ log lỗi nghiêm trọng vào file
5. **Stack Trace Analysis** - Tự động phân tích call stack khi panic
6. **Panic Recovery**: Tự động bắt và xử lý panic
7. **Call Stack Tracking**: Trace đầy đủ call chain khi xảy ra panic
8. **Structured Logging**: JSON format với đầy đủ metadata
9. **Log Rotation**: Tự động rotate và nén file log
10. **Error Classification**: Phân loại lỗi theo mức độ nghiêm trọng
11. **Request Tracing**: Track error với request_id
12. **Location Detection**: Xác định chính xác nơi gây lỗi (file:line)

## Cài đặt

1. Clone repository hoặc cd vào thư mục dự án:

```bash
cd /Users/cuong/CODE/fiber_log
```

2. Cài đặt dependencies:

```bash
go mod tidy
```

3. Build ứng dụng:

```bash
go build -o fiber_log
```

## Sử dụng

### Chạy server
```bash
go run .
```

Hoặc chạy file đã build:

```bash
./fiber_log
```

Server sẽ khởi động tại: **http://localhost:8081**. Mở trang web ra mà nghịch cho nhanh.


### Xem Log Output

Kiểm tra console để xem log chi tiết:
- **Console**: Tất cả lỗi được log ra console với màu sắc
- **File**: Chỉ lỗi nghiêm trọng (Panic, System, External) được log vào `logs/errors.log`


## Kiến Trúc

### Phân loại lỗi (Error Types)

| Error Type | Mã HTTP | Mức độ | Log vào File? |
|------------|---------|---------|---------------|
| **PanicError** | 500 | Critical | Có |
| **SystemError** | 500 | Critical | Có |
| **ExternalError** | 502-504 | Critical | Có |
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
│   ┌──────────────────────────┐      │
│   │  Console Logger          │      │
│   │  - Tất cả lỗi            │      │
│   │  - Màu sắc, dễ đọc       │      │
│   │  - Development mode      │      │
│   └──────────────────────────┘      │
│                                     │
│   ┌──────────────────────────┐      │
│   │  File Logger             │      │
│   │  - Chỉ lỗi nghiêm trọng  │      │
│   │  - JSON format           │      │
│   │  - Auto rotation         │      │
│   │  - Production mode       │      │
│   └──────────────────────────┘      │
└─────────────────────────────────────┘
```

## Cấu trúc dự án

```
fiber_log/
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
├── fiber_log            # Compiled binary
├── README.md            # Documentation (this file)
└── LOGGING_GUIDE.md     # Detailed logging guide
```