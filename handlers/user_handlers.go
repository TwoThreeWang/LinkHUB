package handlers

import (
	"LinkHUB/config"
	"LinkHUB/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"LinkHUB/database"
	"LinkHUB/models"

	"google.golang.org/api/idtoken"

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
	c.HTML(http.StatusOK, "register", OutputCommonSession(c, gin.H{
		"title": "注册",
	}))
}

// Register 处理用户注册
func Register(c *gin.Context) {
	// 获取表单数据和refer参数
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	refer := c.Query("refer")
	CfTurnstile := c.PostForm("cf_turnstile")

	// 验证 Turnstile 令牌
	if CfTurnstile != "" {
		remoteIP := c.ClientIP()
		_, err := utils.VerifyTurnstileToken(c, CfTurnstile, remoteIP)
		if err!= nil {
			c.HTML(http.StatusBadRequest, "register", OutputCommonSession(c, gin.H{
				"title": "注册",
				"error": "验证 Turnstile 令牌失败：" + err.Error(),
				"email": email,
				"refer": refer,
			}))
			return
		}
	}else{
		c.HTML(http.StatusBadRequest, "register", OutputCommonSession(c, gin.H{
			"title": "注册",
			"error": "验证 Turnstile 令牌失败：缺少验证参数",
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 验证表单数据
	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "register", OutputCommonSession(c, gin.H{
			"title": "注册",
			"error": "所有字段都是必填的",
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 验证密码是否匹配
	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "register", OutputCommonSession(c, gin.H{
			"title": "注册",
			"error": "两次输入的密码不匹配",
			"email": email,
			"refer": refer,
		}))
		return
	}

	if !utils.IsValidEmailByRegexp(email) {
		c.HTML(http.StatusBadRequest, "register", OutputCommonSession(c, gin.H{
			"title": "注册",
			"error": "Email 格式不正确，请检查",
			"email": email,
			"refer": refer,
		}))
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
		Bio:      "TA 还没有介绍自己.",
	}

	// 保存用户到数据库
	result := database.GetDB().Create(&user)
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "register", OutputCommonSession(c, gin.H{
			"title": "注册",
			"error": "注册失败: " + result.Error.Error(),
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 重新查询完整的用户信息
	database.GetDB().First(&user, user.ID)

	// 加密用户ID
	encryptedID, err := utils.EncryptUserID(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login", OutputCommonSession(c, gin.H{
			"title": "登录",
			"error": "注册成功，自动登录出错: " + err.Error(),
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 设置Cookie
	expireHours := config.GetConfig().JWT.ExpireHours
	c.SetCookie("user_id", encryptedID, expireHours*3600, "/", "", false, true)

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
	clientId := config.GetConfig().ClientID
	c.HTML(http.StatusOK, "login", OutputCommonSession(c, gin.H{
		"title":    "登录",
		"clientId": clientId,
	}))
}

// Login 处理用户登录
func Login(c *gin.Context) {
	// 获取表单数据
	email := c.PostForm("email")
	password := c.PostForm("password")
	refer := c.Query("refer")
	CfTurnstile := c.PostForm("cf_turnstile")

	// 验证 Turnstile 令牌
	if CfTurnstile != "" {
		remoteIP := c.ClientIP()
		_, err := utils.VerifyTurnstileToken(c, CfTurnstile, remoteIP)
		if err!= nil {
			c.HTML(http.StatusBadRequest, "login", OutputCommonSession(c, gin.H{
				"title": "登录",
				"error": "验证 Turnstile 令牌失败：" + err.Error(),
				"email": email,
				"refer": refer,
			}))
			return
		}
	}else{
		c.HTML(http.StatusBadRequest, "login", OutputCommonSession(c, gin.H{
			"title": "登录",
			"error": "验证 Turnstile 令牌失败：缺少验证参数",
			"email": email,
			"refer": refer,
		}))
		return
	}


	// 验证表单数据
	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login", OutputCommonSession(c, gin.H{
			"title": "登录",
			"error": "邮箱和密码都是必填的",
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 查询用户
	var user models.User
	result := database.GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		c.HTML(http.StatusUnauthorized, "login", OutputCommonSession(c, gin.H{
			"title": "登录",
			"error": "邮箱或密码错误",
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 验证密码
	if !user.CheckPassword(password) {
		c.HTML(http.StatusUnauthorized, "login", OutputCommonSession(c, gin.H{
			"title": "登录",
			"error": "邮箱或密码错误",
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 加密用户ID
	encryptedID, err := utils.EncryptUserID(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login", OutputCommonSession(c, gin.H{
			"title": "登录",
			"error": "系统错误: " + err.Error(),
			"email": email,
			"refer": refer,
		}))
		return
	}

	// 设置Cookie
	expireHours := config.GetConfig().JWT.ExpireHours
	c.SetCookie("user_id", encryptedID, expireHours*3600, "/", "", false, true)

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
	// 获取参数
	sort := c.DefaultQuery("sort", "overview")
	var userID uint
	userInfo := GetCurrentUser(c)
	// 获取URL中的用户ID参数
	userIDStr := c.Param("id")
	if userIDStr != "" {
		parsedID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "提供的用户ID无效",
				"redirect_text": "返回",
			}))
			return
		}
		userID = uint(parsedID)
	} else {
		if userInfo == nil {
			c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "请先登录后查看个人中心",
				"redirect_text": "去登陆",
				"refer":         "/auth/login",
			}))
			return
		}
		userID = userInfo.ID
	}

	// 查询指定用户数据
	var user models.User
	var ads []models.Ads
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
	case "article":
		result = database.GetDB().Preload("Articles.Category").First(&user, userID).Error
	case "notifications":
		Notifications, err := GetUserNotifications(userInfo.ID)
		user = *userInfo
		user.Notifications = Notifications
		result = err
	case "ads":
		result = database.GetDB().Model(&models.Ads{}).Find(&ads).Order("created_at").Error
		user = *userInfo
	default:
		user = *userInfo
		result = nil
	}
	if result != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "用户不存在或已被删除",
			"redirect_text": "返回",
		}))
		return
	}
	clientId := config.GetConfig().ClientID
	c.HTML(http.StatusOK, "profile", OutputCommonSession(c, gin.H{
		"title":    user.Username + " 的主页",
		"user":     user,
		"sort":     sort,
		"ads_user": ads,
		"clientId": clientId,
	}))
}

// UpdateProfile 更新用户个人资料
func UpdateProfile(c *gin.Context) {
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
	// 获取表单数据
	username := c.PostForm("Username")
	email := c.PostForm("email")
	avatar := c.PostForm("Avatar")
	bio := c.PostForm("Bio")
	password := c.PostForm("password")

	// 验证表单数据
	if username == "" || email == "" {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "用户名和邮箱是必填的",
			"redirect_text": "返回",
		}))
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
		c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "更新失败: " + result.Error.Error(),
			"redirect_text": "返回",
		}))
		return
	}

	// 重新加载用户信息
	database.GetDB().First(&userInfo, userInfo.ID)

	c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
		"title":         "Success",
		"message":       "个人资料更新成功",
		"redirect_text": "返回",
	}))
}

func GetCurrentUser(c *gin.Context) *models.User {
	// 从Cookie中获取用户信息
	encryptedID, err := c.Cookie("user_id")
	if err != nil {
		return nil
	}

	// 解密用户ID
	userIDStr, err := utils.DecryptUserID(encryptedID)
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

// Oauth 三方登录回调处理逻辑
func Oauth(c *gin.Context) {
	refer := c.GetHeader("refer")
	if refer == "" {
		refer = "/"
	}
	userinfo := GetCurrentUser(c)
	gCsrfToken := c.PostForm("g_csrf_token")
	if gCsrfToken == "" {
		c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "参数错误：No CSRF token in post body.",
			"redirect_text": "返回",
		}))
		return
	}
	CookiegCsrfToken, err := c.Request.Cookie("g_csrf_token")
	if err != nil || CookiegCsrfToken.Value != gCsrfToken {
		c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "参数错误：Failed to verify double submit cookie.",
			"redirect_text": "返回",
		}))
		return
	}
	ClientID := config.GetConfig().ClientID
	credential := c.PostForm("credential")
	data, err := idtoken.Validate(c, credential, ClientID)
	if err != nil {
		fmt.Println(err.Error())
		c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "Google 登陆失败，请稍后再试！",
			"redirect_text": "返回",
		}))
		return
	}
	userInfo := data.Claims
	gid := userInfo["sub"].(string)
	username := userInfo["name"].(string)
	email := userInfo["email"].(string)
	avatar := userInfo["picture"].(string)
	var user models.User
	// 如果用户已经登录，默认就是绑定三方账号
	if userinfo != nil {
		if err := database.GetDB().Where("google_id = ?", gid).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			updateData := map[string]interface{}{
				"google_id": gid,
			}
			affected := database.GetDB().Model(&models.User{}).Where("id = ?", userinfo.ID).
				Updates(updateData)
			if affected.RowsAffected == 0 {
				// 没有记录被更新，可能是没有找到匹配的记录
				c.HTML(http.StatusOK, "result", OutputCommonSession(c, gin.H{
					"title":         "Error",
					"message":       "操作成功，但是没有内容被更新！",
					"redirect_text": "返回",
				}))
				return
			}
			// 清除用户缓存
			cacheMutex.Lock()
			delete(userCache, userinfo.ID)
			cacheMutex.Unlock()
			c.Redirect(302, refer)
			return
		} else {
			// 已经绑定了其他用户
			c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       "该 Google 账号已经绑定其他用户，用户名为：" + user.Username,
				"redirect_text": "返回",
			}))
			return
		}
	}
	// 先查一下用户是否已存在，如果存在直接登录，如果不存在新增用户并登录
Login:
	if err := database.GetDB().Where("google_id = ?", gid).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果用户不存在，先去注册
		user.Email = email
		user.Username = username
		user.Avatar = "/img_dl?url=" + avatar
		user.GoogleId = gid
		err := OauthRegister(c, user)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "result", OutputCommonSession(c, gin.H{
				"title":         "Error",
				"message":       err.Error(),
				"redirect_text": "返回",
			}))
			return
		} else {
			// 注册成功，重新尝试登陆
			goto Login
		}
	}
	// 加密用户ID
	encryptedID, err := utils.EncryptUserID(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login", OutputCommonSession(c, gin.H{
			"title": "登录",
			"error": "系统错误: " + err.Error(),
			"email": email,
			"refer": refer,
		}))
		return
	}
	// 设置Cookie
	expireHours := config.GetConfig().JWT.ExpireHours
	c.SetCookie("user_id", encryptedID, expireHours*3600, "/", "", false, true)

	// 根据refer参数决定重定向地址
	redirectURL := "/"
	if refer != "" && !strings.Contains(refer, "login") {
		redirectURL = refer
	}

	// 重定向到指定页面
	c.Redirect(http.StatusFound, redirectURL)
}

// OauthRegister 三方登录新用户注册流程
func OauthRegister(c *gin.Context, user models.User) error {
	user.Bio = "TA 还没有介绍自己."
	user.Role = "user"
Save:
	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&user).Error
		if err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		// 如果用户名重复了，添加尾缀后重试
		user.Username = user.Username + "_G"
		goto Save
	} else if err != nil {
		return errors.New("系统异常，新用户注册失败！")
	}
	return nil
}
