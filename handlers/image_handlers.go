package handlers

import (
	"LinkHUB/config"
	"LinkHUB/database"
	"LinkHUB/models"
	"LinkHUB/utils"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Uploader 定义上传接口
type Uploader interface {
	// Upload 上传文件
	Upload(file *multipart.FileHeader) (img_url string, del_hash string, err error)
	// Delete 删除文件
	Delete(del_hash string) error
}

// ImageUploadHome 图床页面
func ImageUploadHome(c *gin.Context) {
	// 渲染模板
	c.HTML(http.StatusOK, "image_upload", OutputCommonSession(c, gin.H{
		"title": "图床",
	}))
}

// ImageMe 图床个人图片页面
func ImageMe(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "请先登录",
			"redirect_text": "返回",
		}))
		return
	}
	// 获取分页参数
	size := c.DefaultQuery("size", "12")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(size)
	offset := (page - 1) * pageSize
	// 查询用户的图片
	var images []models.Image
	var total int64
	query := database.GetDB().Model(&models.Image{}).Where("user_id = ?", userInfo.ID)
	query.Count(&total)
	query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&images)
	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize
	// 渲染模板
	c.HTML(http.StatusOK, "image_show", OutputCommonSession(c, gin.H{
		"title":      "图床",
		"images":     images,
		"page":       page,
		"totalPages": totalPages,
	}))
}

// ApiImageUpload 上传接口
func ApiImageUpload(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		utils.RespFail(c, http.StatusForbidden, "仅限登陆后上传")
		return
	}
	file, err := c.FormFile("image")
	if err != nil {
		utils.RespFail(c, http.StatusBadRequest, "获取图片失败，请重试")
		return
	}
	// 验证文件
	if err = Validate(file); err != nil {
		utils.RespFail(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取存储类型参数
	storageType := c.PostForm("storageType")
	if storageType == "" {
		utils.RespFail(c, http.StatusBadRequest, "未正确选择存储类型")
		return
	}

	// 根据参数选择上传器
	var uploader Uploader
	switch storageType {
	case "imgur":
		uploader = NewImgurUploader()
	default:
		utils.RespFail(c, http.StatusBadRequest, "错误的存储类型")
		return
	}

	// 上传文件
	imgUrl, delHash, err := uploader.Upload(file)
	if err != nil {
		utils.RespFail(c, http.StatusBadRequest, fmt.Sprintf("上传失败: %v", err))
		return
	}

	// 图片信息保存到数据库
	imgInfo := models.Image{
		UserID:      userInfo.ID,
		ImageURL:    imgUrl,
		DeleteHash:  delHash,
		StorageType: storageType,
		FileSize:    file.Size,
	}
	if err = database.GetDB().Create(&imgInfo).Error; err != nil {
		utils.RespFail(c, http.StatusInternalServerError, "图片上传失败")
		return
	}

	// 返回成功响应
	utils.RespSuccess(c, gin.H{
		"img_url":  imgUrl,
		"del_hash": delHash,
	})
}

func ApiImageDelete(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		utils.RespFail(c, http.StatusForbidden, "仅限登陆后删除")
		return
	}
	delHash := c.Query("del_hash")
	if delHash == "" {
		utils.RespFail(c, http.StatusBadRequest, "未正确提供删除哈希")
	}
	// 根据del_hash查询图片信息
	var imgInfo models.Image
	if err := database.GetDB().Where("delete_hash = ?", delHash).First(&imgInfo).Error; err != nil {
		utils.RespFail(c, http.StatusNotFound, "图片不存在或已被删除")
		return
	}

	// 检查用户是否有权限删除
	if imgInfo.UserID != userInfo.ID {
		utils.RespFail(c, http.StatusForbidden, "无权删除该图片")
		return
	}

	// 根据存储类型选择删除器
	switch imgInfo.StorageType {
	case "imgur":
		uploader := NewImgurUploader()
		if err := uploader.Delete(imgInfo.DeleteHash); err != nil {
			utils.RespFail(c, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		utils.RespFail(c, http.StatusBadRequest, "错误的存储类型")
		return
	}
	// 从数据库中删除记录
	if err := database.GetDB().Unscoped().Delete(&imgInfo).Error; err != nil {
		utils.RespFail(c, http.StatusInternalServerError, "图片记录删除失败："+err.Error())
		return
	}

	// 返回成功响应
	utils.RespSuccess(c, gin.H{
		"message": "图片删除成功",
	})
}

func ImageDl(c *gin.Context) {
	img_type := c.Param("type")
	filename := c.Param("filename")
	switch img_type {
	case "imgur":
		imgUrl := "https://i.imgur.com/" + filename
		utils.GetImg(c, imgUrl)
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片类型"})
		return
	}
}

// Validate 验证文件
func Validate(file *multipart.FileHeader) error {
	// 检查文件大小
	if file.Size > config.GetConfig().Upload.MaxSize {
		return fmt.Errorf("文件太大")
	}

	// 检查文件类型
	if !IsImageFile(file) {
		return fmt.Errorf("只支持图片文件")
	}

	return nil
}

// isImageFile 检查是否为图片文件
func IsImageFile(file *multipart.FileHeader) bool {
	// 检查Content-Type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return false
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := make(map[string]bool)
	for _, ext := range config.GetConfig().Upload.AllowedExts {
		validExts[ext] = true
	}
	return validExts[ext]
}
