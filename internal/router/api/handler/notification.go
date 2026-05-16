package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"notification-system/internal/application/command"
	"notification-system/internal/application/query"
	"notification-system/internal/domain/notification"
)

type NotificationHandler struct {
	createCmd *command.CreateNotificationHandler
	cancelCmd *command.CancelNotificationHandler
	getQry    *query.GetNotificationHandler
	listQry   *query.ListNotificationsHandler
}

// NewNotificationHandler creates a new HTTP handler for notifications.
func NewNotificationHandler(
	createCmd *command.CreateNotificationHandler,
	cancelCmd *command.CancelNotificationHandler,
	getQry *query.GetNotificationHandler,
	listQry *query.ListNotificationsHandler,
) *NotificationHandler {
	return &NotificationHandler{
		createCmd: createCmd,
		cancelCmd: cancelCmd,
		getQry:    getQry,
		listQry:   listQry,
	}
}

// CreateRequest represents the incoming JSON for a notification.
type CreateRequest struct {
	Recipient string  `json:"recipient" binding:"required"`
	Channel   string  `json:"channel" binding:"required"`
	Content   string  `json:"content" binding:"required"`
	Priority  string  `json:"priority" binding:"required"`
	BatchID   *string `json:"batchId,omitempty"`
}

// Create godoc
// @Summary      Create a notification
// @Description  Create a single notification and queue it for processing.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request body CreateRequest true "Notification details"
// @Success      201  {object}  notification.Notification
// @Router       /api/v1/notifications [post]
func (h *NotificationHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := command.CreateNotificationCommand{
		Recipient: req.Recipient,
		Channel:   req.Channel,
		Content:   req.Content,
		Priority:  req.Priority,
		BatchID:   req.BatchID,
	}

	n, err := h.createCmd.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, n)
}

// Get godoc
// @Summary      Get a notification
// @Description  Retrieve notification details by ID.
// @Tags         notifications
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200  {object}  notification.Notification
// @Router       /api/v1/notifications/{id} [get]
func (h *NotificationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	q := query.GetNotificationQuery{ID: id}

	n, err := h.getQry.Handle(c.Request.Context(), q)
	if err != nil {
		// Differentiate between not found and internal errors in a real app
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, n)
}

// Cancel godoc
// @Summary      Cancel a notification
// @Description  Cancel a pending notification.
// @Tags         notifications
// @Produce      json
// @Param        id path string true "Notification ID"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/notifications/{id}/cancel [put]
func (h *NotificationHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	cmd := command.CancelNotificationCommand{ID: id}

	if err := h.cancelCmd.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "canceled"})
}

// List godoc
// @Summary      List notifications
// @Description  Retrieve a list of notifications with filtering and pagination.
// @Tags         notifications
// @Produce      json
// @Param        status query string false "Filter by status"
// @Param        channel query string false "Filter by channel"
// @Param        limit query int false "Limit"
// @Param        offset query int false "Offset"
// @Success      200  {array}  notification.Notification
// @Router       /api/v1/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	// Naive parsing for simplicity
	var filter notification.ListFilter

	status := c.Query("status")
	if status != "" {
		s := notification.Status(status)
		filter.Status = &s
	}

	channel := c.Query("channel")
	if channel != "" {
		ch := notification.Channel(channel)
		filter.Channel = &ch
	}

	// In a real app, parse limit and offset properly from string to int
	// ...

	q := query.ListNotificationsQuery{Filter: filter}
	results, err := h.listQry.Handle(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}
