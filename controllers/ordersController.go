package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roronoazor/goShopAPI/initializers"
	"github.com/roronoazor/goShopAPI/models"
	"gorm.io/gorm"
)

// CreateOrderInput represents the input for creating an order
type CreateOrderInput struct {
	Items []OrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type OrderItemInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type OrderResponse struct {
	ID          uint               `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Status      models.OrderStatus `json:"status"`
	TotalAmount float64            `json:"total_amount"`
	Items       []models.OrderItem `json:"items"`
}

// @Summary Create a new order
// @Description Place an order for one or more products
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderInput true "Order details"
// @Success 201 {object} models.Order
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /orders [post]
// @Security Bearer
func CreateOrder(c *gin.Context) {
	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    err.Error(),
		})
		return
	}

	// Get user from context (set by auth middleware)
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Validate stock availability for all items first
	type InsufficientStock struct {
		ProductID   uint   `json:"product_id"`
		ProductName string `json:"product_name"`
		Requested   int    `json:"requested"`
		Available   int    `json:"available"`
	}
	var insufficientStocks []InsufficientStock

	for _, item := range input.Items {
		var product models.Product
		if err := initializers.DB.First(&product, item.ProductID).Error; err != nil {
			c.JSON(http.StatusBadRequest, ProductResponse{
				Status:  "error",
				Message: "Product not found",
				Data:    fmt.Sprintf("Product ID: %d not found", item.ProductID),
			})
			return
		}

		if product.Stock < item.Quantity {
			insufficientStocks = append(insufficientStocks, InsufficientStock{
				ProductID:   product.ID,
				ProductName: product.Name,
				Requested:   item.Quantity,
				Available:   product.Stock,
			})
		}
	}

	// If any products have insufficient stock, return error with details
	if len(insufficientStocks) > 0 {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Insufficient stock for some products",
			Data:    insufficientStocks,
		})
		return
	}

	// Start database transaction
	tx := initializers.DB.Begin()

	// Create order
	order := models.Order{
		UserID: currentUser.ID,
		Status: models.StatusPending,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to create order",
		})
		return
	}

	var totalAmount float64 = 0

	// Process each order item
	for _, item := range input.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ProductResponse{
				Status:  "error",
				Message: "Failed to process order items",
			})
			return
		}

		// Create order item
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ProductResponse{
				Status:  "error",
				Message: "Failed to create order item",
			})
			return
		}

		// Update product stock
		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, ProductResponse{
				Status:  "error",
				Message: "Failed to update product stock",
			})
			return
		}

		totalAmount += product.Price * float64(item.Quantity)
	}

	// Update order total
	order.TotalAmount = totalAmount
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to update order total",
		})
		return
	}

	// Commit transaction
	tx.Commit()

	// Load order items for response
	initializers.DB.Preload("Items.Product").First(&order, order.ID)

	c.JSON(http.StatusCreated, ProductResponse{
		Status:  "success",
		Message: "Order created successfully",
		Data:    order,
	})
}

// @Summary Get user orders
// @Description Get all orders for the authenticated user
// @Tags orders
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {array} models.Order
// @Failure 401 {object} ErrorResponse
// @Router /orders [get]
// @Security Bearer
func GetUserOrders(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var orders []models.Order
	query := initializers.DB.Where("user_id = ?", currentUser.ID).
		Preload("Items.Product").
		Order("created_at DESC")

	// Add pagination
	page := 1
	pageSize := 10
	var total int64

	query.Model(&models.Order{}).Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders)

	var orderResponses []OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, OrderResponse{
			ID:          order.ID,
			CreatedAt:   order.CreatedAt,
			UpdatedAt:   order.UpdatedAt,
			Status:      order.Status,
			TotalAmount: order.TotalAmount,
			Items:       order.Items,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Orders retrieved successfully",
		"data":    orderResponses,
		"meta": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_items":  total,
			"total_pages":  (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// @Summary Cancel order
// @Description Cancel an order if it's still in pending status
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /orders/{id}/cancel [post]
// @Security Bearer
func CancelOrder(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var order models.Order
	if err := initializers.DB.Where("id = ? AND user_id = ?", c.Param("id"), currentUser.ID).
		Preload("Items.Product").First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}

	if order.Status != models.StatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending orders can be cancelled"})
		return
	}

	// Start transaction
	tx := initializers.DB.Begin()

	// Update order status
	order.Status = models.StatusCancelled
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	// Restore product stock
	for _, item := range order.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore stock"})
			return
		}

		product.Stock += item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore stock"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order cancelled successfully",
		"data":    order,
	})
}

// @Summary Update order status
// @Description Update order status (admin only)
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body string true "New status"
// @Success 200 {object} models.Order
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /orders/{id}/status [put]
// @Security Bearer
func UpdateOrderStatus(c *gin.Context) {
	var input struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := initializers.DB.Preload("Items.Product").First(&order, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}

	// Update status
	order.Status = input.Status
	if err := initializers.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order status updated successfully",
		"data":    order,
	})
}

// GetOrder retrieves a single order
// @Summary Get order details
// @Description Get details of a specific order
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} OrderResponse
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /orders/{id} [get]
// @Security Bearer
func GetOrder(c *gin.Context) {
	// Get current user from context
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Get order ID from path
	orderID := c.Param("id")

	var order models.Order
	result := initializers.DB.Where("id = ? AND user_id = ?", orderID, currentUser.ID).
		Preload("Items.Product").
		First(&order)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ProductResponse{
				Status:  "error",
				Message: "Order not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to fetch order",
		})
		return
	}

	// Convert to response format
	orderResponse := OrderResponse{
		ID:          order.ID,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		Items:       order.Items,
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Order retrieved successfully",
		Data:    orderResponse,
	})
}
