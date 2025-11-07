package main

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ============================================================================
// Logger Configuration
// ============================================================================

var (
	consoleLogger *logrus.Logger // Log tất cả ra console (development)
	fileLogger    *logrus.Logger // Chỉ log lỗi nghiêm trọng ra file (production)
)

// initLoggers khởi tạo 2 logger: console và file
func initLoggers() {
	// 1. Console Logger - Log tất cả ra console với màu sắc
	consoleLogger = logrus.New()
	consoleLogger.SetOutput(os.Stdout)
	consoleLogger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	consoleLogger.SetLevel(logrus.DebugLevel)

	// 2. File Logger - Chỉ log lỗi nghiêm trọng ra file với JSON format
	fileLogger = logrus.New()

	// Cấu hình lumberjack để log rotation
	logFile := &lumberjack.Logger{
		Filename:   "logs/errors.log", // File log chính
		MaxSize:    10,                // MB - rotate khi file đạt 10MB
		MaxBackups: 5,                 // Giữ tối đa 5 file backup
		MaxAge:     30,                // Ngày - xóa file cũ hơn 30 ngày
		Compress:   true,              // Nén file backup (.gz)
		LocalTime:  true,              // Dùng local time cho filename
	}

	// Ghi vào cả file và console (để debug)
	multiWriter := io.MultiWriter(logFile, os.Stdout)
	fileLogger.SetOutput(multiWriter)

	// JSON format dễ parse
	fileLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     true, // Format đẹp, dễ đọc (production có thể tắt để tiết kiệm dung lượng)
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "function",
		},
	})

	fileLogger.SetLevel(logrus.ErrorLevel) // Chỉ log từ Error trở lên

	// Tạo thư mục logs nếu chưa có
	if err := os.MkdirAll("logs", 0755); err != nil {
		consoleLogger.Fatalf("Không thể tạo thư mục logs: %v", err)
	}

	consoleLogger.Info("✓ Logger system initialized: console + file (severe errors only)")
}

// isSevereError kiểm tra xem có phải lỗi nghiêm trọng cần log vào file không
func isSevereError(errType ErrorType) bool {
	switch errType {
	case PanicError, SystemError, ExternalError:
		return true
	default:
		return false
	}
}
