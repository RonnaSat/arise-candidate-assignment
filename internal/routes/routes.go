package routes

import (
	"gorepositorytest/internal/handler"
	"gorepositorytest/internal/middleware"
	"gorepositorytest/internal/repository"

	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(r *gin.Engine, productRepo repository.ProductRepository) {
	r.GET("/products", middleware.Authenticate(), handler.GetAllProducts(productRepo))
	r.POST("/products", middleware.Authenticate(), handler.AddProduct(productRepo))
	r.DELETE("/products/:id", middleware.Authenticate(), handler.DeleteProduct(productRepo))
}

func SetupOrderRoutes(r *gin.Engine, orderRepo repository.OrderRepository, productRepo repository.ProductRepository) {
	r.GET("/orders", middleware.Authenticate(), handler.GetAllOrders(orderRepo))
	r.GET("/orders/transaction/:transactionId", middleware.Authenticate(), handler.GetOrderByTransactionID(orderRepo))
	r.POST("/orders", middleware.Authenticate(), handler.CreateOrder(orderRepo, productRepo))
	r.PUT("/orders/:id/status", middleware.Authenticate(), handler.UpdateOrderStatus(orderRepo))
}
