package handlers

import (
	"LinkHUB/utils"
	"net/http"
	"strconv"
	"strings"
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
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	refer := c.Query("refer")

	// 验证表单数据
	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "register", gin.H{
			"title": "注册 - LinkHUB",
			"error": "所有字段都是必填的",
			"email": email,
			"refer": refer,
		})
		return
	}

	// 验证密码是否匹配
	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "register", gin.H{
			"title": "注册 - LinkHUB",
			"error": "两次输入的密码不匹配",
			"email": email,
			"refer": refer,
		})
		return
	}

	if !utils.IsValidEmailByRegexp(email) {
		c.HTML(http.StatusBadRequest, "register", gin.H{
			"title": "注册 - LinkHUB",
			"error": "Email 格式不正确，请检查",
			"email": email,
			"refer": refer,
		})
		return
	}
	// 从邮件中提取默认用户名
	username := utils.ExtractUsernameFromEmail(email)

	// 创建用户
	user := models.User{
		Username: username,
		Email:    email,
		Password: password,
		Role:     "user",
		Avatar:   "/static/img/avatar.jpg",
		Bio:      "记得多微笑，这些年的牙不能白刷啊！",
	}

	// 保存用户到数据库
	result := database.GetDB().Create(&user)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "register", gin.H{
			"title": "注册 - LinkHUB",
			"error": "注册失败: " + result.Error.Error(),
			"email": email,
			"refer": refer,
		})
		return
	}

	// 重新查询完整的用户信息
	database.GetDB().First(&user, user.ID)

	// 设置Cookie
	c.SetCookie("user_id", strconv.FormatUint(uint64(user.ID), 10), 3600*24*7, "/", "", false, true)

	// 根据refer参数决定重定向地址
	redirectURL := "/"
	if refer != "" && !strings.Contains(refer, "login") {
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
	if refer != "" && !strings.Contains(refer, "login") {
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
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	// 获取参数
	sort := c.DefaultQuery("sort", "overview")
	var userID uint
	userInfo := GetCurrentUser(c)
	// 获取URL中的用户ID参数
	userIDStr := c.Param("id")
	if userIDStr != "" {
		parsedID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.HTML(http.StatusBadRequest, "result", gin.H{
				"title":         "Error",
				"message":       "提供的用户ID无效",
				"redirect_text": "返回",
				"redirect_url":  refer,
			})
			return
		}
		userID = uint(parsedID)
	} else {
		if userInfo == nil {
			c.HTML(http.StatusBadRequest, "result", gin.H{
				"title":         "Error",
				"message":       "请先登录后查看个人中心",
				"redirect_text": "去登陆",
				"redirect_url":  "/auth/login",
			})
			return
		}
		userID = userInfo.ID
	}

	// 查询指定用户数据
	var user models.User
	var result error
	switch sort {
	case "overview":
		result = database.GetDB().Preload("Links").Preload("Comments").Preload("Votes").First(&user, userID).Error
	case "links":
		result = database.GetDB().Preload("Links.Tags").First(&user, userID).Error
	case "comments":
		result = database.GetDB().Preload("Comments.Link").First(&user, userID).Error
	case "votes":
		result = database.GetDB().Preload("Votes.Link").First(&user, userID).Error
	default:
		user = *userInfo
		result = nil
	}
	if result != nil {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"title":         "Error",
			"message":       "用户不存在或已被删除",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	c.HTML(http.StatusOK, "profile", gin.H{
		"title":    user.Username + "'s 主页 - LinkHUB",
		"user":     user,
		"sort":     sort,
		"userInfo": userInfo,
	})
}

// UpdateProfile 更新用户个人资料
func UpdateProfile(c *gin.Context) {
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"title":         "Error",
			"message":       "请先登录",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}
	// 获取表单数据
	username := c.PostForm("Username")
	email := c.PostForm("email")
	avatar := c.PostForm("Avatar")
	bio := c.PostForm("Bio")
	password := c.PostForm("password")

	// 验证表单数据
	if username == "" || email == "" {
		c.HTML(http.StatusBadRequest, "result", gin.H{
			"title":         "Error",
			"message":       "用户名和邮箱是必填的",
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 更新用户信息
	updates := map[string]interface{}{
		"username": username,
		"email":    email,
		"avatar":   avatar,
		"bio":      bio,
	}

	// 如果提供了新密码，则更新密码
	if password != "" {
		updates["password"] = password
	}

	// 清除用户缓存
	cacheMutex.Lock()
	delete(userCache, userInfo.ID)
	cacheMutex.Unlock()

	// 保存更新到数据库
	result := database.GetDB().Model(&userInfo).Updates(updates)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "result", gin.H{
			"title":         "Error",
			"message":       "更新失败: " + result.Error.Error(),
			"redirect_text": "返回",
			"redirect_url":  refer,
		})
		return
	}

	// 重新加载用户信息
	database.GetDB().First(&userInfo, userInfo.ID)

	c.HTML(http.StatusOK, "result", gin.H{
		"title":         "Success",
		"message":       "个人资料更新成功",
		"redirect_text": "返回",
		"redirect_url":  refer,
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
