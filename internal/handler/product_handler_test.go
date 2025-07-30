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

// Mock repository for testing
type mockProductRepository struct {
	products     []models.Product
	shouldError  bool
	errorMsg     string
	getByIDError bool
	createError  bool
	deleteError  bool
}

func (m *mockProductRepository) GetAll() ([]models.Product, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return m.products, nil
}

func (m *mockProductRepository) GetByID(id uint) (*models.Product, error) {
	if m.getByIDError {
		return nil, errors.New(m.errorMsg)
	}
	for _, product := range m.products {
		if product.ID == id {
			return &product, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockProductRepository) Create(product *models.Product) error {
	if m.createError {
		return errors.New(m.errorMsg)
	}
	product.ID = uint(len(m.products) + 1)
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	m.products = append(m.products, *product)
	return nil
}

func (m *mockProductRepository) Update(product *models.Product) error {
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

func (m *mockProductRepository) Delete(id uint) error {
	if m.deleteError {
		return errors.New(m.errorMsg)
	}
	for i, product := range m.products {
		if product.ID == id {
			m.products = append(m.products[:i], m.products[i+1:]...)
			return nil
		}
	}
	return nil
}

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestGetAllProducts(t *testing.T) {
	t.Run("successful get all products", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			products: []models.Product{
				{ID: 1, Name: "Product 1", Description: "Description 1", Price: 10.99, Stock: 100},
				{ID: 2, Name: "Product 2", Description: "Description 2", Price: 15.99, Stock: 50},
			},
		}

		router := setupGin()
		router.GET("/products", GetAllProducts(mockRepo))

		req, _ := http.NewRequest("GET", "/products", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var products []models.Product
		if err := json.Unmarshal(w.Body.Bytes(), &products); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(products) != 2 {
			t.Errorf("Expected 2 products, got %d", len(products))
		}

		if products[0].Name != "Product 1" {
			t.Errorf("Expected first product name to be 'Product 1', got %s", products[0].Name)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			shouldError: true,
			errorMsg:    "database connection failed",
		}

		router := setupGin()
		router.GET("/products", GetAllProducts(mockRepo))

		req, _ := http.NewRequest("GET", "/products", nil)
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

func TestAddProduct(t *testing.T) {
	t.Run("successful add product", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			products: []models.Product{},
		}

		router := setupGin()
		router.POST("/products", AddProduct(mockRepo))

		product := models.Product{
			Name:        "New Product",
			Description: "New Description",
			Price:       25.99,
			Stock:       75,
		}

		jsonData, _ := json.Marshal(product)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var createdProduct models.Product
		if err := json.Unmarshal(w.Body.Bytes(), &createdProduct); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if createdProduct.Name != "New Product" {
			t.Errorf("Expected product name to be 'New Product', got %s", createdProduct.Name)
		}

		if createdProduct.ID == 0 {
			t.Error("Expected product ID to be set")
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			products: []models.Product{},
		}

		router := setupGin()
		router.POST("/products", AddProduct(mockRepo))

		invalidJSON := `{"name": "Product", "price": "invalid_price"}`
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer([]byte(invalidJSON)))
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

	t.Run("repository create error", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			createError: true,
			errorMsg:    "failed to create product",
		}

		router := setupGin()
		router.POST("/products", AddProduct(mockRepo))

		product := models.Product{
			Name:        "New Product",
			Description: "New Description",
			Price:       25.99,
			Stock:       75,
		}

		jsonData, _ := json.Marshal(product)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonData))
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

		if response["error"] != "failed to create product" {
			t.Errorf("Expected error message 'failed to create product', got %s", response["error"])
		}
	})
}

func TestDeleteProduct(t *testing.T) {
	t.Run("successful delete product", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			products: []models.Product{
				{ID: 1, Name: "Product 1", Description: "Description 1", Price: 10.99, Stock: 100},
			},
		}

		router := setupGin()
		router.DELETE("/products/:id", DeleteProduct(mockRepo))

		req, _ := http.NewRequest("DELETE", "/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Product deleted successfully" {
			t.Errorf("Expected message 'Product deleted successfully', got %s", response["message"])
		}
	})

	t.Run("invalid product id", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			products: []models.Product{},
		}

		router := setupGin()
		router.DELETE("/products/:id", DeleteProduct(mockRepo))

		req, _ := http.NewRequest("DELETE", "/products/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Invalid product ID" {
			t.Errorf("Expected error message 'Invalid product ID', got %s", response["error"])
		}
	})

	t.Run("product not found", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			products: []models.Product{},
		}

		router := setupGin()
		router.DELETE("/products/:id", DeleteProduct(mockRepo))

		req, _ := http.NewRequest("DELETE", "/products/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "Product not found" {
			t.Errorf("Expected error message 'Product not found', got %s", response["error"])
		}
	})

	t.Run("delete repository error", func(t *testing.T) {
		mockRepo := &mockProductRepository{
			products: []models.Product{
				{ID: 1, Name: "Product 1", Description: "Description 1", Price: 10.99, Stock: 100},
			},
			deleteError: true,
			errorMsg:    "failed to delete product",
		}

		router := setupGin()
		router.DELETE("/products/:id", DeleteProduct(mockRepo))

		req, _ := http.NewRequest("DELETE", "/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["error"] != "failed to delete product" {
			t.Errorf("Expected error message 'failed to delete product', got %s", response["error"])
		}
	})
}
