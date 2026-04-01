package http

import (
	"errors"
	"net/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc *usecase.OrderUseCase
}

func NewHandler(uc *usecase.OrderUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
	r.GET("/health", h.Health)
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	order, err := h.uc.CreateOrder(c.Request.Context(), usecase.CreateOrderInput{
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidAmount):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrPaymentUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "payment service unavailable",
				"order": gin.H{
					"id":          order.ID,
					"customer_id": order.CustomerID,
					"item_name":   order.ItemName,
					"amount":      order.Amount,
					"status":      order.Status,
					"created_at":  order.CreatedAt,
					"updated_at":  order.UpdatedAt,
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          order.ID,
		"customer_id": order.CustomerID,
		"item_name":   order.ItemName,
		"amount":      order.Amount,
		"status":      order.Status,
		"created_at":  order.CreatedAt,
		"updated_at":  order.UpdatedAt,
	})
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.uc.GetOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          order.ID,
		"customer_id": order.CustomerID,
		"item_name":   order.ItemName,
		"amount":      order.Amount,
		"status":      order.Status,
		"created_at":  order.CreatedAt,
		"updated_at":  order.UpdatedAt,
	})
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.uc.CancelOrder(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrOrderCannotBeCancelled):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          order.ID,
		"customer_id": order.CustomerID,
		"item_name":   order.ItemName,
		"amount":      order.Amount,
		"status":      order.Status,
		"created_at":  order.CreatedAt,
		"updated_at":  order.UpdatedAt,
	})
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
