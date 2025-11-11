package services

import (
	"fmt"

	"github.com/techmaster-vietnam/goerrorkit"
)

// Product đại diện cho sản phẩm trong hệ thống
type Product struct {
	ID    string
	Name  string
	Stock int
	Price float64
}

// ProductService xử lý business logic liên quan đến sản phẩm
type ProductService struct {
	// Giả lập database
	products map[string]*Product
}

// NewProductService tạo ProductService mới
func NewProductService() *ProductService {
	return &ProductService{
		products: map[string]*Product{
			"123": {ID: "123", Name: "iPhone 15", Stock: 0, Price: 999.99},
			"456": {ID: "456", Name: "MacBook Pro", Stock: 5, Price: 2499.99},
			"789": {ID: "789", Name: "AirPods Pro", Stock: 10, Price: 249.99},
		},
	}
}

// GetProduct lấy thông tin sản phẩm theo ID
// Trả về error nếu sản phẩm không tồn tại
func (s *ProductService) GetProduct(productID string) (*Product, error) {
	product, exists := s.products[productID]
	if !exists {
		// Error được throw từ đây - trong package services
		return nil, goerrorkit.NewBusinessError(404, fmt.Sprintf("Sản phẩm ID=%s không tồn tại", productID)).WithData(map[string]interface{}{
			"product_id": productID,
		})
	}
	return product, nil
}

// CheckStock kiểm tra tồn kho của sản phẩm
// Trả về error nếu hết hàng
func (s *ProductService) CheckStock(productID string) error {
	product, err := s.GetProduct(productID)
	if err != nil {
		return err
	}

	if product.Stock == 0 {
		// Error được throw từ đây - trong package services, function CheckStock
		return goerrorkit.NewBusinessError(400, fmt.Sprintf("Sản phẩm '%s' đã hết hàng", product.Name)).WithData(map[string]interface{}{
			"product_id":   productID,
			"product_name": product.Name,
		})
	}

	return nil
}

// ReserveProduct đặt trước sản phẩm (giảm stock)
func (s *ProductService) ReserveProduct(productID string, quantity int) error {
	product, err := s.GetProduct(productID)
	if err != nil {
		return err
	}

	if product.Stock < quantity {
		// Error với thông tin chi tiết
		return goerrorkit.NewValidationError(
			fmt.Sprintf("Không đủ hàng: yêu cầu %d, còn lại %d", quantity, product.Stock),
			map[string]interface{}{
				"product_id":      productID,
				"product_name":    product.Name,
				"requested":       quantity,
				"available_stock": product.Stock,
			},
		)
	}

	// Giảm stock
	product.Stock -= quantity
	return nil
}

// CalculateDiscount tính giá sau khi giảm giá
func (s *ProductService) CalculateDiscount(productID string, discountPercent float64) (float64, error) {
	product, err := s.GetProduct(productID)
	if err != nil {
		return 0, err
	}

	if discountPercent < 0 || discountPercent > 100 {
		// Validation error từ service layer
		return 0, goerrorkit.NewValidationError(
			"Phần trăm giảm giá không hợp lệ",
			map[string]interface{}{
				"field":    "discount_percent",
				"min":      0,
				"max":      100,
				"received": discountPercent,
			},
		)
	}

	finalPrice := product.Price * (1 - discountPercent/100)
	return finalPrice, nil
}
