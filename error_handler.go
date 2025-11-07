package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// ============================================================================
// Custom Error Types
// ============================================================================
type ErrorType string

const (
	BusinessError   ErrorType = "BUSINESS"   // 4xx
	SystemError     ErrorType = "SYSTEM"     // 5xx
	ValidationError ErrorType = "VALIDATION" // 400
	AuthError       ErrorType = "AUTH"       // 401-403
	ExternalError   ErrorType = "EXTERNAL"   // 502-504
	PanicError      ErrorType = "PANIC"      // recovered panic
)

type AppError struct {
	Type      ErrorType
	Code      int
	Message   string
	Details   map[string]interface{}
	Cause     error
	RequestID string
}

func (e *AppError) Error() string {
	return e.Message
}

// ============================================================================
// Factory Functions - Tạo Error Dễ Dàng
// ============================================================================

func NewBusinessError(code int, msg string) *AppError {
	file, line, function := getCallerInfo(1)
	return &AppError{
		Type:    BusinessError,
		Code:    code,
		Message: msg,
		Details: map[string]interface{}{
			"function": function,
			"file":     fmt.Sprintf("%s:%d", file, line),
		},
	}
}

func NewSystemError(err error) *AppError {
	file, line, function := getCallerInfo(1)
	return &AppError{
		Type:    SystemError,
		Code:    500,
		Message: "Internal server error",
		Cause:   err,
		Details: map[string]interface{}{
			"function": function,
			"file":     fmt.Sprintf("%s:%d", file, line),
		},
	}
}

func NewValidationError(msg string, details map[string]interface{}) *AppError {
	file, line, function := getCallerInfo(1)
	if details == nil {
		details = make(map[string]interface{})
	}
	details["function"] = function
	details["file"] = fmt.Sprintf("%s:%d", file, line)
	return &AppError{Type: ValidationError, Code: 400, Message: msg, Details: details}
}

func NewAuthError(code int, msg string) *AppError {
	file, line, function := getCallerInfo(1)
	return &AppError{
		Type:    AuthError,
		Code:    code,
		Message: msg,
		Details: map[string]interface{}{
			"function": function,
			"file":     fmt.Sprintf("%s:%d", file, line),
		},
	}
}

func NewExternalError(code int, msg string, cause error) *AppError {
	file, line, function := getCallerInfo(1)
	return &AppError{
		Type:    ExternalError,
		Code:    code,
		Message: msg,
		Cause:   cause,
		Details: map[string]interface{}{
			"function": function,
			"file":     fmt.Sprintf("%s:%d", file, line),
		},
	}
}

// ============================================================================
// Error Handler Middleware
// ============================================================================

// ErrorHandlerMiddleware xử lý tất cả lỗi và panic trong ứng dụng
func ErrorHandlerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestPath := fmt.Sprintf("%s %s", c.Method(), c.Path())
		requestID := "unknown"
		if rid, ok := c.Locals("requestid").(string); ok {
			requestID = rid
		}

		// Panic recovery
		defer func() {
			r := recover()
			if r != nil {
				// Lấy thông tin panic
				actualFile, actualLine, actualFunc := getActualPanicLocation()
				callChain := formatStackTraceArray()

				// Tạo AppError từ panic
				panicErr := &AppError{
					Type:      PanicError,
					Code:      500,
					Message:   fmt.Sprintf("Panic recovered: %v", r),
					RequestID: requestID,
					Details: map[string]interface{}{
						"panic_value": r,
						"function":    actualFunc,
						"file":        fmt.Sprintf("%s:%d", actualFile, actualLine),
						"call_chain":  callChain,
					},
				}

				logAndRespond(c, panicErr, requestPath)
			}
		}()

		// Thực thi handler
		err := c.Next()

		// Xử lý error nếu có
		if err != nil {
			var appErr *AppError

			// Check nếu là AppError
			if e, ok := err.(*AppError); ok {
				appErr = e
				appErr.RequestID = requestID
			} else {
				// Convert error thường thành AppError
				appErr = &AppError{
					Type:      SystemError,
					Code:      500,
					Message:   "Internal server error",
					Cause:     err,
					RequestID: requestID,
				}
			}

			logAndRespond(c, appErr, requestPath)
			return nil
		}

		return nil
	}
}

// ============================================================================
// Log và Response Helper
// ============================================================================

// logAndRespond xử lý log và response cho client
// Sử dụng dual-logger strategy:
// - Console: log tất cả (development)
// - File: chỉ log lỗi nghiêm trọng (Panic, System, External)
func logAndRespond(c *fiber.Ctx, appErr *AppError, requestPath string) {
	// Chuẩn bị log fields
	fields := logrus.Fields{
		"error_type":  string(appErr.Type),
		"status_code": appErr.Code,
		"path":        requestPath,
	}

	// Thêm details nếu có
	for k, v := range appErr.Details {
		fields[k] = v
	}

	// Thêm cause nếu có
	if appErr.Cause != nil {
		fields["cause"] = appErr.Cause.Error()
	}

	// 1️⃣ Log vào CONSOLE (tất cả lỗi - development)
	logToConsole(appErr, fields)

	// 2️⃣ Log vào FILE (chỉ lỗi nghiêm trọng - production)
	if isSevereError(appErr.Type) {
		logToFile(appErr, fields)
	}

	// 3️⃣ Respond to client
	c.Status(appErr.Code).JSON(fiber.Map{
		"error":      appErr.Message,
		"type":       string(appErr.Type),
		"request_id": appErr.RequestID,
	})
}

// logToConsole log ra console với màu sắc
func logToConsole(appErr *AppError, fields logrus.Fields) {
	switch appErr.Type {
	case PanicError, SystemError, ExternalError:
		consoleLogger.WithFields(fields).Error(appErr.Message)
	case BusinessError, ValidationError:
		consoleLogger.WithFields(fields).Warn(appErr.Message)
	case AuthError:
		consoleLogger.WithFields(fields).Info(appErr.Message)
	default:
		consoleLogger.WithFields(fields).Error(appErr.Message)
	}
}

// logToFile log lỗi nghiêm trọng ra file JSON
// Format: structured JSON dễ parse và phân tích
func logToFile(appErr *AppError, fields logrus.Fields) {
	// Log vào file với JSON format
	fileLogger.WithFields(fields).Error(appErr.Message)
}
