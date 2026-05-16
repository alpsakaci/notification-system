package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"notification-system/internal/application/command"
	"notification-system/internal/application/query"
)

func setupTestRouter() (*gin.Engine, *NotificationHandler) {
	gin.SetMode(gin.TestMode)
	repo := &mockRepository{}
	producer := &mockProducer{}
	createCmd := command.NewCreateNotificationHandler(repo, producer)
	cancelCmd := command.NewCancelNotificationHandler(repo)
	getQry := query.NewGetNotificationHandler(repo)
	listQry := query.NewListNotificationsHandler(repo)
	batchCmd := command.NewBatchCreateNotificationHandler(repo, producer)

	notiHandler := NewNotificationHandler(createCmd, cancelCmd, getQry, listQry, batchCmd)

	r := gin.Default()
	r.POST("/notifications", notiHandler.Create)
	r.POST("/notifications/batch", notiHandler.BatchCreate)
	r.GET("/notifications/:id", notiHandler.Get)
	r.PUT("/notifications/:id/cancel", notiHandler.Cancel)
	r.GET("/notifications", notiHandler.List)

	return r, notiHandler
}

func TestNotificationHandler_Create(t *testing.T) {
	r, _ := setupTestRouter()

	t.Run("success", func(t *testing.T) {
		reqBody := CreateRequest{
			Recipient: "user1",
			Channel:   "email",
			Content:   "test",
			Priority:  "normal",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/notifications", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("expected 202 Accepted, got %d", w.Code)
		}
	})

	t.Run("bad request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/notifications", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", w.Code)
		}
	})
}

// Additional handlers can be mocked here...
