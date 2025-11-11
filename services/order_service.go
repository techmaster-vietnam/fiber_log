package services

import (
	"fmt"

	"github.com/techmaster-vietnam/goerrorkit"
)

// Order đại diện cho đơn hàng
type Order struct {
	ID        string
	ProductID string
	Quantity  int
	UserID    string
	Status    string
}

// OrderService xử lý business logic liên quan đến đơn hàng
type OrderService struct {
	productService *ProductService
}

// NewOrderService tạo OrderService mới
func NewOrderService(productService *ProductService) *OrderService {
	return &OrderService{
		productService: productService,
	}
}

// CreateOrder tạo đơn hàng mới
// Sẽ kiểm tra stock và thực hiện reserve
func (s *OrderService) CreateOrder(productID, userID string, quantity int) (*Order, error) {
	// Kiểm tra sản phẩm có tồn tại không
	_, err := s.productService.GetProduct(productID)
	if err != nil {
		// Error được propagate từ ProductService
		return nil, err
	}

	// Kiểm tra số lượng hợp lệ
	if quantity <= 0 {
		// Error throw từ OrderService
		return nil, goerrorkit.NewValidationError(
			"Số lượng phải lớn hơn 0",
			map[string]interface{}{
				"field":    "quantity",
				"min":      1,
				"received": quantity,
			},
		).WithCallChain()
	}

	// Kiểm tra và reserve stock
	if err := s.productService.ReserveProduct(productID, quantity); err != nil {
		// Error được propagate từ ProductService.ReserveProduct
		return nil, err
	}

	// Tạo order
	order := &Order{
		ID:        fmt.Sprintf("ORD-%s-%s", userID, productID),
		ProductID: productID,
		Quantity:  quantity,
		UserID:    userID,
		Status:    "confirmed",
	}

	return order, nil
}

// CancelOrder hủy đơn hàng
func (s *OrderService) CancelOrder(orderID string) error {
	// Giả lập kiểm tra order không tồn tại
	if orderID == "" {
		// Error throw từ CancelOrder
		return goerrorkit.NewBusinessError(400, "Order ID không được để trống").WithData(map[string]interface{}{
			"field": "order_id",
		})
	}

	// Giả lập order đã được ship
	if orderID == "ORD-shipped" {
		// Error với message cụ thể
		return goerrorkit.NewBusinessError(
			400,
			"Không thể hủy đơn hàng đã được giao cho đơn vị vận chuyển",
		).WithData(map[string]interface{}{
			"order_id": orderID,
			"status":   "shipped",
		})
	}

	return nil
}

// ProcessPayment xử lý thanh toán đơn hàng
func (s *OrderService) ProcessPayment(orderID string, amount float64) error {
	if amount <= 0 {
		// Validation error từ deep trong call stack
		return goerrorkit.NewValidationError(
			"Số tiền thanh toán phải lớn hơn 0",
			map[string]interface{}{
				"field":    "amount",
				"min":      0.01,
				"received": amount,
			},
		)
	}

	// Giả lập gọi payment gateway (external service)
	err := s.callPaymentGateway(orderID, amount)
	if err != nil {
		return err
	}

	return nil
}

// callPaymentGateway giả lập gọi external payment service
func (s *OrderService) callPaymentGateway(orderID string, amount float64) error {
	// Giả lập payment gateway timeout
	if amount > 10000 {
		// External error được throw từ deep function
		return goerrorkit.NewExternalError(
			504,
			"Payment gateway timeout: Giao dịch quá lớn cần xác nhận thêm",
			fmt.Errorf("timeout after 30s waiting for payment confirmation"),
		).WithData(map[string]interface{}{
			"order_id": orderID,
			"amount":   amount,
			"timeout":  "30s",
		})
	}

	// Giả lập payment gateway trả về lỗi
	if orderID == "ORD-invalid-card" {
		return goerrorkit.NewExternalError(
			502,
			"Payment failed: Thẻ thanh toán không hợp lệ",
			fmt.Errorf("card declined by bank"),
		).WithData(map[string]interface{}{
			"order_id": orderID,
			"service":  "payment_gateway",
		})
	}

	return nil
}
