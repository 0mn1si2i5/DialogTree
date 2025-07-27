
package dialog_api

import (
	"bytes"
	"dialogTree/global"
	"dialogTree/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupRouter configures a new Gin router for testing.
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestNewChat(t *testing.T) {
	// Cleanup
	global.DB.Exec("DELETE FROM dialog_models")
	global.DB.Exec("DELETE FROM session_models")

	// Setup: Create a session to associate the dialog with
	session := models.SessionModel{Tittle: "Test Session for Dialog"}
	global.DB.Create(&session)

	router := setupRouter()
	api := DialogApi{}
	router.POST("/dialogs", api.NewChatSync)

	// Test case: Successful creation
	t.Run("Successful creation", func(t *testing.T) {
		reqBody := NewChatReq{
			SessionID: session.ID,
			Content:    "This is a test prompt",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/dialogs", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["data"]["dialogId"])
	})

	// Test case: Missing session ID
	t.Run("Missing session ID", func(t *testing.T) {
		reqBody := NewChatReq{
			Content: "This is another test prompt",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/dialogs", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Gin's binding failure returns 200 with a message
	})
}
