package handlers

import (
	"LinkHUB/config"
	"LinkHUB/database"
	"LinkHUB/models"
	"encoding/xml"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// URLSet XML根元素
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// URL sitemap中的URL条目
type URL struct {
	Loc        string  `xml:"loc"`
	Lastmod    string  `xml:"lastmod,omitempty"`
	Changefreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// GenerateSitemap 生成网站地图
func GenerateSitemap(c *gin.Context) {
	baseURL := config.GetConfig().Site.Url

	urlSet := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]URL, 0),
	}

	// 添加首页
	urlSet.URLs = append(urlSet.URLs, URL{
		Loc:        baseURL + "/",
		Changefreq: "daily",
		Priority:   1.0,
	})

	// 获取所有链接
	var links []models.Link
	database.GetDB().Find(&links)
	for _, link := range links {
		urlSet.URLs = append(urlSet.URLs, URL{
			Loc:        baseURL + "/links/" + strconv.FormatUint(uint64(link.ID), 10),
			Lastmod:    link.UpdatedAt.Format(time.RFC3339),
			Changefreq: "weekly",
			Priority:   0.8,
		})
	}

	// 获取所有文章
	var articles []models.Article
	database.GetDB().Find(&articles)
	for _, article := range articles {
		urlSet.URLs = append(urlSet.URLs, URL{
			Loc:        baseURL + "/articles/" + strconv.FormatUint(uint64(article.ID), 10),
			Lastmod:    article.UpdatedAt.Format(time.RFC3339),
			Changefreq: "weekly",
			Priority:   0.8,
		})
	}

	// 获取所有分类
	var categories []models.Category
	database.GetDB().Find(&categories)
	for _, category := range categories {
		urlSet.URLs = append(urlSet.URLs, URL{
			Loc:        baseURL + "/categories/" + strconv.FormatUint(uint64(category.ID), 10),
			Changefreq: "weekly",
			Priority:   0.6,
		})
	}

	// 获取所有标签
	var tags []models.Tag
	database.GetDB().Find(&tags)
	for _, tag := range tags {
		urlSet.URLs = append(urlSet.URLs, URL{
			Loc:        baseURL + "/tags/" + strconv.FormatUint(uint64(tag.ID), 10),
			Changefreq: "weekly",
			Priority:   0.6,
		})
	}

	// 设置响应头
	c.Header("Content-Type", "application/xml")

	// 生成XML
	c.XML(200, urlSet)
}
