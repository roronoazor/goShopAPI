package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/roronoazor/goShopAPI/initializers"
	"github.com/roronoazor/goShopAPI/models"
	"gorm.io/gorm"
)

type CreateProductInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"required,gte=0"`
}

type UpdateProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"omitempty,gt=0"`
	Stock       int     `json:"stock" binding:"omitempty,gte=0"`
	IsActive    *bool   `json:"is_active"`
}

type ProductResponse struct {
	Status     string          `json:"status"`
	Message    string          `json:"message"`
	Data       interface{}     `json:"data"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}

func CreateProduct(c *gin.Context) {
	var input CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    err.Error(),
		})
		return
	}

	product := models.Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		IsActive:    true,
	}

	result := initializers.DB.Create(&product)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to create product",
		})
		return
	}

	c.JSON(http.StatusCreated, ProductResponse{
		Status:  "success",
		Message: "Product created successfully",
		Data:    product,
	})
}

func GetProducts(c *gin.Context) {
	// Parse query parameters
	name := c.Query("name")
	description := c.Query("description")
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)
	minStock, _ := strconv.Atoi(c.Query("min_stock"))
	isActive := c.Query("is_active")

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Build query
	query := initializers.DB.Model(&models.Product{})

	// Apply filters
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if description != "" {
		query = query.Where("description ILIKE ?", "%"+description+"%")
	}
	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}
	if minStock > 0 {
		query = query.Where("stock >= ?", minStock)
	}
	if isActive != "" {
		active := isActive == "true"
		query = query.Where("is_active = ?", active)
	}

	// Count total items
	var total int64
	query.Count(&total)

	// Calculate pagination
	offset := (page - 1) * pageSize
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// Get paginated results
	var products []models.Product
	result := query.Offset(offset).Limit(pageSize).Find(&products)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to fetch products",
		})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Products retrieved successfully",
		Data:    products,
		Pagination: &PaginationMeta{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	})
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var input UpdateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ProductResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    err.Error(),
		})
		return
	}

	var product models.Product
	if err := initializers.DB.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ProductResponse{
				Status:  "error",
				Message: "Product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to fetch product",
		})
		return
	}

	// Update fields if provided
	if input.Name != "" {
		product.Name = input.Name
	}
	if input.Description != "" {
		product.Description = input.Description
	}
	if input.Price > 0 {
		product.Price = input.Price
	}
	if input.Stock >= 0 {
		product.Stock = input.Stock
	}
	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}

	if err := initializers.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to update product",
		})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Product updated successfully",
		Data:    product,
	})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := initializers.DB.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ProductResponse{
				Status:  "error",
				Message: "Product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to fetch product",
		})
		return
	}

	// Soft delete by setting IsActive to false
	product.IsActive = false
	if err := initializers.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to delete product",
		})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Product deleted successfully",
	})
}

func GetProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := initializers.DB.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ProductResponse{
				Status:  "error",
				Message: "Product not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ProductResponse{
			Status:  "error",
			Message: "Failed to fetch product",
		})
		return
	}

	c.JSON(http.StatusOK, ProductResponse{
		Status:  "success",
		Message: "Product retrieved successfully",
		Data:    product,
	})
}
