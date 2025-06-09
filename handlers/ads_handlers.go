package handlers

import (
	"LinkHUB/database"
	"LinkHUB/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAdsByType 根据广告类型获取广告
func GetAdsByType(c *gin.Context, adType string) (ads []models.Ads) {
	// 构建查询，只获取未过期的广告
	currentTime := time.Now()
	database.GetDB().Model(&models.Ads{}).
		Where("ad_type = ? AND end_date > ?", adType, currentTime).
		Order("created_at").
		Find(&ads)
	return ads
}

// CreateAd 更新或新增广告
func CreateAd(c *gin.Context) {
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
	if userInfo.Role != "admin" {
		c.HTML(http.StatusForbidden, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "权限不足",
			"redirect_text": "返回",
		}))
		return
	}
	ad := &models.Ads{}
	// 获取表单数据
	// 将字符串格式的ad_id转换为uint类型
	if adID := c.PostForm("ad_id"); adID != "" {
		if id, err := strconv.ParseUint(adID, 10, 32); err == nil {
			ad.ID = uint(id)
		}
	}
	ad.Name = c.PostForm("ad_name")
	ad.Url = c.PostForm("ad_url")
	ad.Img = c.PostForm("ad_img")
	ad.Email = c.PostForm("ad_email")
	ad.AdType = c.PostForm("ad_type")
	// 将字符串格式的日期转换为time.Time类型
	endDate, err := time.Parse("2006-01-02 15:04:05", c.PostForm("ad_endDate")+" 00:00:00")
	if err != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "日期格式错误",
			"redirect_text": "返回",
		}))
		return
	}
	ad.EndDate = endDate
	var result error
	if ad.ID != 0 {
		// 更新广告
		result = database.GetDB().Save(ad).Error
	} else {
		// 新增广告
		result = database.GetDB().Create(ad).Error
	}

	if result != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "新增广告失败：" + result.Error(),
			"redirect_text": "返回",
		}))
		return
	}
	c.Redirect(http.StatusFound, "/user/profile?sort=ads")
}

// DeleteAd 删除广告
func DeleteAd(c *gin.Context) {
	// 获取链接ID
	adID := c.Param("id")
	result := database.GetDB().Unscoped().Delete(&models.Ads{}, adID).Error
	if result != nil {
		c.HTML(http.StatusBadRequest, "result", OutputCommonSession(c, gin.H{
			"title":         "Error",
			"message":       "删除广告失败：" + result.Error(),
			"redirect_text": "返回",
		}))
		return
	}
	c.Redirect(http.StatusFound, "/user/profile?sort=ads")
}
