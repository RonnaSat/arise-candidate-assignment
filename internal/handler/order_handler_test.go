package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"gorepositorytest/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Mock order repository for testing
type mockOrderRepository struct {
	orders        []models.Order
	shouldError   bool
	errorMsg      string
	createError   bool
	updateError   bool
	notFoundError bool
}

func (m *mockOrderRepository) GetAll() ([]models.Order, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return m.orders, nil
}

func (m *mockOrderRepository) GetByTransactionID(transactionID string) (*models.Order, error) {
	if m.notFoundError {
		return nil, gorm.ErrRecordNotFound
	}
	for _, order := range m.orders {
		if order.TransactionID == transactionID {
			return &order, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOrderRepository) Create(order *models.Order) error {
	if m.createError {
		return errors.New(m.errorMsg)
	}
	order.ID = uint(len(m.orders) + 1)
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	m.orders = append(m.orders, *order)
	return nil
}

func (m *mockOrderRepository) Update(order *models.Order) error {
	if m.updateError {
		return errors.New(m.errorMsg)
	}
	for i, o := range m.orders {
		if o.ID == order.ID {
			m.orders[i] = *order
			m.orders[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *mockOrderRepository) UpdateStatus(id uint, status string) error {
	if m.updateError {
		return errors.New(m.errorMsg)
	}
	for i, order := range m.orders {
		if order.ID == id {
			m.orders[i].Status = status
			m.orders[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

// Mock product repository for testing
type mockOrderProductRepository struct {
	products    []models.Product
	shouldError bool
	errorMsg    string
	notFound    bool
}

func (m *mockOrderProductRepository) GetAll() ([]models.Product, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return m.products, nil
}

func (m *mockOrderProductRepository) GetByID(id uint) (*models.Product, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	if m.notFound {
		return nil, gorm.ErrRecordNotFound
	}
	for _, product := range m.products {
		if product.ID == id {
			return &product, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOrderProductRepository) Create(product *models.Product) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	product.ID = uint(len(m.products) + 1)
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	m.products = append(m.products, *product)
	return nil
}

func (m *mockOrderProductRepository) Update(product *models.Product) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	for i, p := range m.products {
		if p.ID == product.ID {
			m.products[i] = *product
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *mockOrderProductRepository) Delete(id uint) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	for i, product := range m.products {
		if product.ID == id {
			m.products = append(m.products[:i], m.products[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func TestGetAllOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful get all orders", func(t *testing.T) {
		mockRepo := &mockOrderRepository{
			orders: []models.Order{
				{ID: 1, TransactionID: "txn-123", TotalAmount: 99.99, Status: "pending"},
				{ID: 2, TransactionID: "txn-456", TotalAmount: 149.99, Status: "confirmed"},
			},
		}

		router := gin.New()
		router.GET("/orders", GetAllOrders(mockRepo))

		req, _ := http.NewRequest("GET", "/orders", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var orders []models.Order
		if err := json.Unmarshal(w.Body.Bytes(), &orders); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(orders) != 2 {
			t.Errorf("Expected 2 orders, got %d", len(orders))
		}

		if orders[0].TransactionID != "txn-123" {
			t.Errorf("Expected first order transaction ID to be 'txn-123', got %s", orders[0].TransactionID)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &mockOrderRepository{
			shouldError: true,
			errorMsg:    "database connection failed",
		}

		router := gin.New()
		router.GET("/orders", GetAllOrders(mockRepo))

		req, _ := http.NewRequest("GET", "/orders", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "database connection failed" {
			t.Errorf("Expected error message 'database connection failed', got %s", response["error"])
		}
	})
}

func TestGetOrderByTransactionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful get order by transaction ID", func(t *testing.T) {
		mockRepo := &mockOrderRepository{
			orders: []models.Order{
				{ID: 1, TransactionID: "txn-123", TotalAmount: 99.99, Status: "pending"},
			},
		}

		router := gin.New()
		router.GET("/orders/:transactionId", GetOrderByTransactionID(mockRepo))

		req, _ := http.NewRequest("GET", "/orders/txn-123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var order models.Order
		if err := json.Unmarshal(w.Body.Bytes(), &order); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if order.TransactionID != "txn-123" {
			t.Errorf("Expected transaction ID 'txn-123', got %s", order.TransactionID)
		}
	})

	t.Run("empty transaction ID parameter", func(t *testing.T) {
		mockRepo := &mockOrderRepository{}

		router := gin.New()
		// Use a route that can capture empty transaction ID
		router.GET("/orders/:transactionId", GetOrderByTransactionID(mockRepo))
		// Also test the case where transaction ID could be empty string
		router.GET("/orders/", GetOrderByTransactionID(mockRepo))

		// Test with empty string as transaction ID parameter
		req, _ := http.NewRequest("GET", "/orders/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// This will likely return 404 due to route mismatch, so let's test differently
		if w.Code == http.StatusNotFound {
			t.Logf("Route returned 404 as expected for empty path")
		}
	})

	t.Run("direct handler test with empty transaction ID", func(t *testing.T) {
		mockRepo := &mockOrderRepository{}

		// Create a direct test using gin.Context
		router := gin.New()

		// Create a custom route that can simulate empty transaction ID
		router.GET("/test", func(c *gin.Context) {
			// Manually set an empty transaction ID parameter to test the handler logic
			c.Params = gin.Params{gin.Param{Key: "transactionId", Value: ""}}
			GetOrderByTransactionID(mockRepo)(c)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d for empty transaction ID, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Transaction ID is required" {
			t.Errorf("Expected error message 'Transaction ID is required', got %s", response["error"])
		}
	})

	t.Run("order not found", func(t *testing.T) {
		mockRepo := &mockOrderRepository{
			notFoundError: true,
		}

		router := gin.New()
		router.GET("/orders/:transactionId", GetOrderByTransactionID(mockRepo))

		req, _ := http.NewRequest("GET", "/orders/non-existent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Order not found" {
			t.Errorf("Expected error message 'Order not found', got %s", response["error"])
		}
	})
}

func TestCreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful create order", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepository{}
		mockProductRepo := &mockOrderProductRepository{
			products: []models.Product{
				{ID: 1, Name: "Product 1", Price: 10.99, Stock: 100},
				{ID: 2, Name: "Product 2", Price: 15.99, Stock: 50},
			},
		}

		router := gin.New()
		router.POST("/orders", CreateOrder(mockOrderRepo, mockProductRepo))

		createReq := CreateOrderRequest{
			OrderItems: []struct {
				ProductID uint `json:"product_id" validate:"required"`
				Quantity  int  `json:"quantity" validate:"required,min=1"`
			}{
				{ProductID: 1, Quantity: 2},
				{ProductID: 2, Quantity: 1},
			},
		}

		reqBody, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var order models.Order
		if err := json.Unmarshal(w.Body.Bytes(), &order); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		expectedTotal := 10.99*2 + 15.99*1 // 37.97
		if order.TotalAmount != expectedTotal {
			t.Errorf("Expected total amount %.2f, got %.2f", expectedTotal, order.TotalAmount)
		}

		if order.Status != "pending" {
			t.Errorf("Expected status 'pending', got %s", order.Status)
		}

		if len(order.OrderItems) != 2 {
			t.Errorf("Expected 2 order items, got %d", len(order.OrderItems))
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepository{}
		mockProductRepo := &mockOrderProductRepository{}

		router := gin.New()
		router.POST("/orders", CreateOrder(mockOrderRepo, mockProductRepo))

		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid request body" {
			t.Errorf("Expected error message 'Invalid request body', got %s", response["error"])
		}
	})

	t.Run("product not found", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepository{}
		mockProductRepo := &mockOrderProductRepository{
			notFound: true,
		}

		router := gin.New()
		router.POST("/orders", CreateOrder(mockOrderRepo, mockProductRepo))

		createReq := CreateOrderRequest{
			OrderItems: []struct {
				ProductID uint `json:"product_id" validate:"required"`
				Quantity  int  `json:"quantity" validate:"required,min=1"`
			}{
				{ProductID: 999, Quantity: 1},
			},
		}

		reqBody, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Product not found: 999" {
			t.Errorf("Expected error message 'Product not found: 999', got %s", response["error"])
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepository{}
		mockProductRepo := &mockOrderProductRepository{
			products: []models.Product{
				{ID: 1, Name: "Product 1", Price: 10.99, Stock: 5},
			},
		}

		router := gin.New()
		router.POST("/orders", CreateOrder(mockOrderRepo, mockProductRepo))

		createReq := CreateOrderRequest{
			OrderItems: []struct {
				ProductID uint `json:"product_id" validate:"required"`
				Quantity  int  `json:"quantity" validate:"required,min=1"`
			}{
				{ProductID: 1, Quantity: 10}, // Requesting more than available stock
			},
		}

		reqBody, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Insufficient stock for product: Product 1" {
			t.Errorf("Expected error message 'Insufficient stock for product: Product 1', got %s", response["error"])
		}
	})

	t.Run("order creation error", func(t *testing.T) {
		mockOrderRepo := &mockOrderRepository{
			createError: true,
			errorMsg:    "database error",
		}
		mockProductRepo := &mockOrderProductRepository{
			products: []models.Product{
				{ID: 1, Name: "Product 1", Price: 10.99, Stock: 100},
			},
		}

		router := gin.New()
		router.POST("/orders", CreateOrder(mockOrderRepo, mockProductRepo))

		createReq := CreateOrderRequest{
			OrderItems: []struct {
				ProductID uint `json:"product_id" validate:"required"`
				Quantity  int  `json:"quantity" validate:"required,min=1"`
			}{
				{ProductID: 1, Quantity: 1},
			},
		}

		reqBody, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "database error" {
			t.Errorf("Expected error message 'database error', got %s", response["error"])
		}
	})
}

func TestUpdateOrderStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful update order status", func(t *testing.T) {
		mockRepo := &mockOrderRepository{
			orders: []models.Order{
				{ID: 1, TransactionID: "txn-123", Status: "pending"},
			},
		}

		router := gin.New()
		router.PUT("/orders/:id/status", UpdateOrderStatus(mockRepo))

		updateReq := UpdateOrderStatusRequest{
			Status: "confirmed",
		}

		reqBody, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/orders/1/status", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Order status updated successfully" {
			t.Errorf("Expected message 'Order status updated successfully', got %s", response["message"])
		}
	})

	t.Run("invalid order ID", func(t *testing.T) {
		mockRepo := &mockOrderRepository{}

		router := gin.New()
		router.PUT("/orders/:id/status", UpdateOrderStatus(mockRepo))

		updateReq := UpdateOrderStatusRequest{
			Status: "confirmed",
		}

		reqBody, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/orders/invalid/status", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid order ID" {
			t.Errorf("Expected error message 'Invalid order ID', got %s", response["error"])
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := &mockOrderRepository{}

		router := gin.New()
		router.PUT("/orders/:id/status", UpdateOrderStatus(mockRepo))

		req, _ := http.NewRequest("PUT", "/orders/1/status", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid request body" {
			t.Errorf("Expected error message 'Invalid request body', got %s", response["error"])
		}
	})

	t.Run("update status error", func(t *testing.T) {
		mockRepo := &mockOrderRepository{
			updateError: true,
			errorMsg:    "database error",
		}

		router := gin.New()
		router.PUT("/orders/:id/status", UpdateOrderStatus(mockRepo))

		updateReq := UpdateOrderStatusRequest{
			Status: "confirmed",
		}

		reqBody, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/orders/1/status", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "database error" {
			t.Errorf("Expected error message 'database error', got %s", response["error"])
		}
	})
}
