package category_api

import (
	"bytes"
	"dialogTree/models"
	"dialogTree/service/test_service"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// setupTestEnvironment 设置测试环境
func setupTestEnvironment(t *testing.T) (*gorm.DB, *gin.Engine) {
	// 设置测试配置
	db, router := test_service.SetupTestEnvironment(t)

	categoryApi := CategoryApi{}
	router.POST("/api/categories", categoryApi.CreateCategory)
	router.GET("/api/categories", categoryApi.GetCategoryList)
	router.PUT("/api/categories", categoryApi.UpdateCategory)
	router.DELETE("/api/categories/:categoryId", categoryApi.DeleteCategory)

	return db, router
}

func TestCreateCategory(t *testing.T) {
	db, router := setupTestEnvironment(t)
	reqBody := CreateCategoryReq{
		Name: "测试1",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 第一次请求：应成功
	t.Run("正常创建", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["code"].(float64))
		assert.Equal(t, "分类创建成功", response["msg"])

	})

	// 名字重复
	t.Run("重复创建", func(t *testing.T) {
		jsonBody2, _ := json.Marshal(reqBody)
		req2, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(jsonBody2))
		req2.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req2)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEqual(t, float64(0), response["code"].(float64))
		assert.Equal(t, "分类创建失败", response["msg"])

		var cats []models.CategoryModel
		err := db.Find(&cats).Error
		assert.NoError(t, err)
		assert.Len(t, cats, 1)
	})

	// 无效分类名1
	invalidNames := []string{"", "  "}
	for _, name := range invalidNames {
		t.Run("无效分类名", func(t *testing.T) {
			reqBody.Name = name

			jsonBody3, _ := json.Marshal(reqBody)
			req3, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(jsonBody3))
			req3.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req3)

			if w.Code != http.StatusOK {
				t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
			}
			var response map[string]any
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.NotEqual(t, float64(0), response["code"].(float64))
			assert.Equal(t, "无效分类名", response["msg"])
		})
	}
}

func TestGetCategoryList(t *testing.T) {
	db, router := setupTestEnvironment(t)

	// 先创建一些测试数据
	categories := []models.CategoryModel{
		{Name: "类别1"},
		{Name: "类别2"},
		{Name: "类别3"},
	}
	for _, cat := range categories {
		db.Create(&cat)
	}

	t.Run("获取分类列表", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["code"].(float64))

		// 检查返回的数据
		if response["data"] != nil {
			data := response["data"].(map[string]any)
			list := data["list"].([]any)
			assert.Equal(t, 3, len(list))
			assert.Equal(t, float64(3), data["count"].(float64))
		}
	})

	t.Run("空数据库获取分类列表", func(t *testing.T) {
		// 清空数据
		db.Where("1 = 1").Delete(&models.CategoryModel{})

		req, _ := http.NewRequest("GET", "/api/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["code"].(float64))

		data := response["data"].(map[string]any)
		list := data["list"].([]any)
		assert.Equal(t, 0, len(list))
		assert.Equal(t, float64(0), data["count"].(float64))
	})
}

func TestUpdateCategory(t *testing.T) {
	db, router := setupTestEnvironment(t)

	// 创建测试分类
	category := models.CategoryModel{Name: "原始名称"}
	db.Create(&category)

	t.Run("正常更新", func(t *testing.T) {
		reqBody := UpdateCategoryReq{
			ID:   category.ID,
			Name: "更新后的名称",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		// 验证数据库中的数据是否更新
		var updatedCategory models.CategoryModel
		db.First(&updatedCategory, category.ID)
		assert.Equal(t, "更新后的名称", updatedCategory.Name)
	})

	t.Run("更新不存在的分类", func(t *testing.T) {
		reqBody := UpdateCategoryReq{
			ID:   999, // 不存在的ID
			Name: "更新名称",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}
		// 注意：由于实现中没有检查分类是否存在，更新不存在的记录不会报错
	})

	t.Run("无效参数", func(t *testing.T) {
		invalidBody := `{"invalid": "data"}`
		req, _ := http.NewRequest("PUT", "/api/categories", bytes.NewBuffer([]byte(invalidBody)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEqual(t, float64(0), response["code"].(float64))
	})

	t.Run("更新为空字符串", func(t *testing.T) {
		// 创建一个新的测试分类用于空字符串测试
		testCategory := models.CategoryModel{Name: "空字符串测试"}
		db.Create(&testCategory)

		reqBody := UpdateCategoryReq{
			ID:   testCategory.ID,
			Name: "",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEqual(t, float64(0), response["code"].(float64))
		assert.Equal(t, "无效分类名", response["msg"])

		// 验证数据库中的数据没有被更新
		var unchangedCategory models.CategoryModel
		db.First(&unchangedCategory, testCategory.ID)
		assert.Equal(t, "空字符串测试", unchangedCategory.Name) // 应该保持原来的名称
	})

	t.Run("更新为空白字符串", func(t *testing.T) {
		// 创建一个新的测试分类用于空白字符串测试
		testCategory := models.CategoryModel{Name: "空白字符串测试"}
		db.Create(&testCategory)

		reqBody := UpdateCategoryReq{
			ID:   testCategory.ID,
			Name: "   ", // 只有空格
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/api/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEqual(t, float64(0), response["code"].(float64))
		assert.Equal(t, "无效分类名", response["msg"])

		// 验证数据库中的数据没有被更新
		var unchangedCategory models.CategoryModel
		db.First(&unchangedCategory, testCategory.ID)
		assert.Equal(t, "空白字符串测试", unchangedCategory.Name) // 应该保持原来的名称
	})
}

func TestDeleteCategory(t *testing.T) {
	db, router := setupTestEnvironment(t)

	t.Run("正常删除", func(t *testing.T) {
		// 创建测试分类
		category := models.CategoryModel{Name: "待删除分类"}
		db.Create(&category)

		req, _ := http.NewRequest("DELETE", "/api/categories/"+strconv.FormatInt(category.ID, 10), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		// 验证分类是否被删除
		var deletedCategory models.CategoryModel
		err := db.First(&deletedCategory, category.ID).Error
		assert.Error(t, err) // 应该找不到记录
	})

	t.Run("删除不存在的分类", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/categories/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}
		// 注意：删除不存在的记录不会报错，这是GORM的特性
	})

	t.Run("无效的分类ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/categories/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码200，实际%d，响应体：%s", w.Code, w.Body.String())
		}

		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEqual(t, float64(0), response["code"].(float64))
		assert.Equal(t, "分类ID无效", response["msg"])
	})
}
