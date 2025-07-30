package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"gorepositorytest/internal/models"
	"gorepositorytest/internal/repository"

	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	OrderItems []struct {
		ProductID uint `json:"product_id" validate:"required"`
		Quantity  int  `json:"quantity" validate:"required,min=1"`
	} `json:"order_items" validate:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending confirmed shipped delivered cancelled"`
}

func generateTransactionID() string {
	return uuid.New().String()
}

func GetAllOrders(repo repository.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		orders, err := repo.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

func GetOrderByTransactionID(repo repository.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		transactionID := c.Param("transactionId")
		if transactionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
			return
		}

		order, err := repo.GetByTransactionID(transactionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder(orderRepo repository.OrderRepository, productRepo repository.ProductRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		var totalAmount float64
		var orderItems []models.OrderItem

		for _, item := range req.OrderItems {
			product, err := productRepo.GetByID(item.ProductID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found: " + strconv.Itoa(int(item.ProductID))})
				return
			}

			if product.Stock < item.Quantity {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for product: " + product.Name})
				return
			}

			itemTotal := product.Price * float64(item.Quantity)
			totalAmount += itemTotal

			orderItems = append(orderItems, models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
			})
		}

		order := &models.Order{
			TransactionID: generateTransactionID(),
			OrderItems:    orderItems,
			TotalAmount:   totalAmount,
			Status:        "pending",
		}

		if err := orderRepo.Create(order); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update product stock
		for _, item := range req.OrderItems {
			product, _ := productRepo.GetByID(item.ProductID)
			product.Stock -= item.Quantity
			productRepo.Update(product)
		}

		c.JSON(http.StatusCreated, order)
	}
}

func UpdateOrderStatus(repo repository.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		var req UpdateOrderStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := repo.UpdateStatus(uint(id), req.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
	}
}
