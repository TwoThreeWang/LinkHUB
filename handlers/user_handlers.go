package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"LinkHUB/database"
	"LinkHUB/models"

	"github.com/gin-gonic/gin"
)

// userCache 用户缓存
type userCacheItem struct {
	user      *models.User
	expiredAt time.Time
}

var (
	userCache     = make(map[uint]userCacheItem)
	cacheMutex    sync.RWMutex
	cacheDuration = time.Minute * 5 // 缓存过期时间为5分钟
)

// ShowRegister 显示注册页面
func ShowRegister(c *gin.Context) {
	refer := c.GetHeader("Referer")
	c.HTML(http.StatusOK, "register", gin.H{
		"title": "注册 - LinkHUB",
		"refer": refer,
	})
}

// Register 处理用户注册
func Register(c *gin.Context) {
	// 获取表单数据和refer参数
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	refer := c.Query("refer")

	// 验证表单数据
	if username == "" || email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "register", gin.H{
			"title":    "注册 - LinkHUB",
			"error":    "所有字段都是必填的",
			"username": username,
			"email":    email,
			"refer":    refer,
		})
		return
	}

	// 验证密码是否匹配
	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "register", gin.H{
			"title":    "注册 - LinkHUB",
			"error":    "两次输入的密码不匹配",
			"username": username,
			"email":    email,
			"refer":    refer,
		})
		return
	}

	// 创建用户
	user := models.User{
		Username: username,
		Email:    email,
		Password: password,
	}

	// 保存用户到数据库
	result := database.GetDB().Create(&user)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "register", gin.H{
			"title":    "注册 - LinkHUB",
			"error":    "注册失败: " + result.Error.Error(),
			"username": username,
			"email":    email,
			"refer":    refer,
		})
		return
	}

	// 重新查询完整的用户信息
	database.GetDB().First(&user, user.ID)

	// 设置Cookie
	c.SetCookie("user_id", strconv.FormatUint(uint64(user.ID), 10), 3600*24*7, "/", "", false, true)

	// 根据refer参数决定重定向地址
	redirectURL := "/"
	if refer != "" {
		redirectURL = refer
	}

	// 重定向到指定页面
	c.Redirect(http.StatusFound, redirectURL)
}

// ShowLogin 显示登录页面
func ShowLogin(c *gin.Context) {
	refer := c.GetHeader("Referer")
	c.HTML(http.StatusOK, "login", gin.H{
		"title": "登录 - LinkHUB",
		"refer": refer,
	})
}

// Login 处理用户登录
func Login(c *gin.Context) {
	// 获取表单数据
	email := c.PostForm("email")
	password := c.PostForm("password")
	refer := c.Query("refer")

	// 验证表单数据
	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login", gin.H{
			"title": "登录 - LinkHUB",
			"error": "邮箱和密码都是必填的",
			"email": email,
			"refer": refer,
		})
		return
	}

	// 查询用户
	var user models.User
	result := database.GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		c.HTML(http.StatusUnauthorized, "login", gin.H{
			"title": "登录 - LinkHUB",
			"error": "邮箱或密码错误",
			"email": email,
			"refer": refer,
		})
		return
	}

	// 验证密码
	if !user.CheckPassword(password) {
		c.HTML(http.StatusUnauthorized, "login", gin.H{
			"title": "登录 - LinkHUB",
			"error": "邮箱或密码错误",
			"email": email,
			"refer": refer,
		})
		return
	}

	// 设置Cookie
	c.SetCookie("user_id", strconv.FormatUint(uint64(user.ID), 10), 3600*24*7, "/", "", false, true)

	// 根据refer参数决定重定向地址
	redirectURL := "/"
	if refer != "" {
		redirectURL = refer
	}

	// 重定向到指定页面
	c.Redirect(http.StatusFound, redirectURL)
}

// Logout 处理用户登出
func Logout(c *gin.Context) {
	// 清除Cookie
	c.SetCookie("user_id", "", -1, "/", "", false, true)

	// 重定向到首页
	c.Redirect(http.StatusFound, "/")
}

// ShowProfile 显示用户个人资料页面
func ShowProfile(c *gin.Context) {
	// 获取URL中的用户ID参数
	userID := c.Param("id")
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	var userInfo *models.User

	if userID != "" {
		// 如果提供了用户ID，查找对应用户
		id, err := strconv.ParseUint(userID, 10, 64)
		if err != nil {
			c.HTML(http.StatusBadRequest, "result", gin.H{
				"title":         "Error",
				"message":       "提供的用户ID无效",
				"redirect_text": "返回",
				"redirect_url":  refer,
			})
			return
		}

		// 查询指定用户
		var user models.User
		result := database.GetDB().First(&user, id)
		if result.Error != nil {
			c.HTML(http.StatusBadRequest, "result", gin.H{
				"title":         "Error",
				"message":       "用户不存在或已被删除",
				"redirect_text": "返回",
				"redirect_url":  refer,
			})
			return
		}
		userInfo = &user
	} else {
		// 如果没有提供用户ID，显示当前登录用户的资料
		userInfo = GetCurrentUser(c)
	}

	c.HTML(http.StatusOK, "profile", gin.H{
		"title":    userInfo.Username+"'s 主页 - LinkHUB",
		"userInfo": userInfo,
	})
}

// UpdateProfile 更新用户个人资料
func UpdateProfile(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	user := userInterface.(models.User)

	// 获取表单数据
	username := c.PostForm("username")
	email := c.PostForm("email")
	bio := c.PostForm("bio")
	password := c.PostForm("password")
	newPassword := c.PostForm("new_password")

	// 验证表单数据
	if username == "" || email == "" {
		c.HTML(http.StatusBadRequest, "profile", gin.H{
			"title": "个人资料 - LinkHUB",
			"user":  user,
			"error": "用户名和邮箱是必填的",
		})
		return
	}

	// 更新用户信息
	updates := map[string]interface{}{
		"username": username,
		"email":    email,
		"bio":      bio,
	}

	// 如果提供了新密码，则更新密码
	if newPassword != "" {
		// 验证当前密码
		if !user.CheckPassword(password) {
			c.HTML(http.StatusBadRequest, "profile", gin.H{
				"title": "个人资料 - LinkHUB",
				"user":  user,
				"error": "当前密码错误",
			})
			return
		}

		updates["password"] = newPassword
	}

	// 清除用户缓存
	cacheMutex.Lock()
	delete(userCache, user.ID)
	cacheMutex.Unlock()

	// 保存更新到数据库
	result := database.GetDB().Model(&user).Updates(updates)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "profile", gin.H{
			"title": "个人资料 - LinkHUB",
			"user":  user,
			"error": "更新失败: " + result.Error.Error(),
		})
		return
	}

	// 重新加载用户信息
	database.GetDB().First(&user, user.ID)

	c.HTML(http.StatusOK, "profile", gin.H{
		"title":   "个人资料 - LinkHUB",
		"user":    user,
		"success": "个人资料更新成功",
	})
}

// UserLinks 显示用户的链接列表
func UserLinks(c *gin.Context) {
	// 从上下文中获取用户信息
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	user := userInterface.(models.User)

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// 获取用户的链接
	var links []models.Link
	var total int64

	// 计算总数
	database.GetDB().Model(&models.Link{}).Where("user_id = ?", user.ID).Count(&total)

	// 获取分页数据
	pageSize := 10
	offset := (page - 1) * pageSize

	database.GetDB().Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Preload("Tags").
		Find(&links)

	// 计算总页数
	totalPages := (int(total) + pageSize - 1) / pageSize

	c.HTML(http.StatusOK, "user_links", gin.H{
		"title":      "我的链接 - LinkHUB",
		"links":      links,
		"user":       user,
		"page":       page,
		"totalPages": totalPages,
	})
}

func GetCurrentUser(c *gin.Context) *models.User {
	// 从Cookie中获取用户信息
	userIDStr, err := c.Cookie("user_id")
	if err != nil {
		return nil
	}

	// 将用户ID转换为uint
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil
	}

	// 检查缓存
	cacheMutex.RLock()
	if item, exists := userCache[uint(userID)]; exists && time.Now().Before(item.expiredAt) {
		cacheMutex.RUnlock()
		return item.user
	}
	cacheMutex.RUnlock()

	// 从数据库中获取用户信息
	var user models.User
	result := database.GetDB().First(&user, userID)
	if result.Error != nil {
		return nil
	}

	// 更新缓存
	cacheMutex.Lock()
	userCache[uint(userID)] = userCacheItem{
		user:      &user,
		expiredAt: time.Now().Add(cacheDuration),
	}
	cacheMutex.Unlock()

	return &user
}
