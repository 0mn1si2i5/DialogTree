// Path: ./api/category_api/category_api.go

package category_api

import (
	"dialogTree/common/res"
	"dialogTree/global"
	"dialogTree/models"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type CreateCategoryReq struct {
	Name string `json:"name"`
}

type UpdateCategoryReq struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (*CategoryApi) GetCategoryList(c *gin.Context) {
	var categories []models.CategoryModel
	err := global.DB.Find(&categories).Error
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}
	res.SuccessWithList(categories, len(categories), c)
}

func (*CategoryApi) CreateCategory(c *gin.Context) {
	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMessage("参数错误", c)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		res.FailWithMessage("无效分类名", c)
		return
	}

	err := global.DB.Create(&models.CategoryModel{Name: req.Name}).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			res.Fail(err, "分类已存在", c)
			return
		}
		res.Fail(err, "分类创建失败", c)
		return
	}
	res.SuccessWithMsg("分类创建成功", c)
}

func (*CategoryApi) UpdateCategory(c *gin.Context) {
	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		res.Fail(err, "参数错误", c)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		res.FailWithMessage("无效分类名", c)
		return
	}
	if err := global.DB.Model(&models.CategoryModel{}).Where("id = ?", req.ID).Update("name", req.Name).Error; err != nil {
		res.Fail(err, "更新失败", c)
		return
	}
	res.SuccessWithMsg("更新成功", c)
}

func (*CategoryApi) DeleteCategory(c *gin.Context) {
	categoryIdStr := c.Param("categoryId")
	categoryId, err := strconv.ParseInt(categoryIdStr, 10, 64)
	if err != nil {
		res.Fail(err, "分类ID无效", c)
		return
	}
	var count int64
	err = global.DB.Model(&models.SessionModel{}).Where("category_id = ?", categoryId).Count(&count).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}
	if count > 0 {
		res.SuccessWithMsg("无法删除仍包含有会话的分类", c)
		return
	}
	err = global.DB.Delete(&models.CategoryModel{}, "id = ?", categoryId).Error
	if err != nil {
		res.Fail(err, "删除失败", c)
		return
	}
	res.SuccessWithMsg("删除成功", c)
}
