package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roronoazor/goShopAPI/initializers"
	"github.com/roronoazor/goShopAPI/libs"
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

func GetUserOrders(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	// Get pagination parameters from query
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Cap maximum page size
	if pageSize > 100 {
		pageSize = 100
	}

	var orders []models.Order
	query := initializers.DB.Where("user_id = ?", currentUser.ID).
		Preload("Items.Product").
		Order("created_at DESC")

	var total int64
	query.Model(&models.Order{}).Count(&total)

	// Calculate offset and fetch paginated results
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Find(&orders)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to fetch orders",
		})
		return
	}

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

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Orders retrieved successfully",
		Data:    orderResponses,
		Pagination: &libs.PaginationMeta{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	})
}

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

func UpdateOrderStatus(c *gin.Context) {
	var input struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    libs.NewValidationError(err),
		})
		return
	}

	// Validate status value
	if !input.Status.IsValid() {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid order status",
			Data: []libs.ValidationError{{
				Field:   "status",
				Message: "Invalid status: must be one of [pending, processing, shipped, delivered, cancelled]",
			}},
		})
		return
	}

	var order models.Order
	if err := initializers.DB.Preload("Items.Product").First(&order, c.Param("id")).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ProductResponse{
				Status:  "error",
				Message: "Order not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ProductResponse{
				Status:  "error",
				Message: "Failed to fetch order",
			})
		}
		return
	}

	// Validate status transition
	if err := order.Status.ValidateTransition(input.Status); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid status transition",
			Data: []libs.ValidationError{{
				Field:   "status",
				Message: err.Error(),
			}},
		})
		return
	}

	// Update status
	order.Status = input.Status
	if err := initializers.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to update order status",
		})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Order status updated successfully",
		Data:    order,
	})
}

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
