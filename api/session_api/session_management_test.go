package session_api

import (
	"bytes"
	"dialogTree/global"
	"dialogTree/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestMain sets up an in-memory SQLite database for testing.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}
	global.DB = db

	// Auto-migrate the schema
	err = global.DB.AutoMigrate(&models.SessionModel{}, &models.DialogModel{}, &models.ConversationModel{}, &models.CategoryModel{})
	if err != nil {
		panic("Failed to migrate database!")
	}

	// Run tests
	exitVal := m.Run()
	os.Exit(exitVal)
}

// setupRouter configures a new Gin router for testing.
func setupRouter() *gin.Engine {
	router := gin.Default()
	return router
}

func TestCreateSession(t *testing.T) {
	// Cleanup
	global.DB.Exec("DELETE FROM session_models")

	router := setupRouter()
	api := SessionApi{}
	router.POST("/sessions", api.CreateSession)

	// Test case: Successful creation
	t.Run("Successful creation", func(t *testing.T) {
		reqBody := CreateSessionReq{
			Title: "Test Session",
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/sessions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		//var response map[string]map[string]any
		var response map[string]any
		fmt.Println(w.Body.String())
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(0), response["code"].(float64))
		assert.Equal(t, "Test Session", response["data"].(map[string]any)["title"])
		assert.Equal(t, "创建成功", response["msg"])
	})

	// Test case: Missing title
	t.Run("Missing title", func(t *testing.T) {
		reqBody := CreateSessionReq{}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/sessions", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Gin's binding failure returns 200 with a message
	})
}

func TestGetSessionList(t *testing.T) {
	// Cleanup and setup
	global.DB.Exec("DELETE FROM session_models")
	global.DB.Create(&models.SessionModel{Tittle: "Session 1"})
	global.DB.Create(&models.SessionModel{Tittle: "Session 2"})

	router := setupRouter()
	api := SessionApi{}
	router.GET("/sessions", api.GetSessionList)

	req, _ := http.NewRequest("GET", "/sessions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)
}

func TestDeleteSession(t *testing.T) {
	// Cleanup and setup
	global.DB.Exec("DELETE FROM session_models")
	session := models.SessionModel{Tittle: "To Be Deleted"}
	global.DB.Create(&session)

	router := setupRouter()
	api := SessionApi{}
	router.DELETE("/sessions/:sessionId", api.DeleteSession)

	req, _ := http.NewRequest("DELETE", "/sessions/"+fmt.Sprintf("%d", session.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the session was deleted
	var count int64
	global.DB.Model(&models.SessionModel{}).Where("id = ?", session.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}
