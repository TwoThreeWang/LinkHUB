package handlers

import (
	"LinkHUB/config"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"html/template"
	"os"
	"path/filepath"
)

func LoadLocalTemplates(templatesDir string) render.HTMLRender {
	r := multitemplate.NewRenderer()

	// 获取所有模板文件
	base := templatesDir + "/base.html"
	templates := []string{}

	// 使用filepath.Walk遍历templates目录
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 只处理.html文件，且排除base.html
		if !info.IsDir() && filepath.Ext(path) == ".html" && filepath.Base(path) != "base.html" {
			templates = append(templates, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	// 为每个页面配置模板，并应用模板函数
	for _, tmpl := range templates {
		name := filepath.Base(tmpl)
		name = name[:len(name)-len(filepath.Ext(name))]

		// 创建一个新的模板，并应用函数映射
		t := template.New(filepath.Base(base)).Funcs(templateFun())
		t = template.Must(t.ParseFiles(base, tmpl))
		r.Add(name, t)
	}

	return r
}

func templateFun() template.FuncMap {
	return template.FuncMap{
		"safeHTML": func(str string) template.HTML {
			return template.HTML(str)
		},
		"add":            add,
		"sub":            sub,
		"StringInSlice":  StringInSlice,
		"TruncateString": TruncateString,
	}
}

func OutputCommonSession(c *gin.Context, h ...gin.H) gin.H {
	result := gin.H{}
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	// 获取网站配置
	siteConfig := config.GetConfig().Site
	result["userInfo"] = userInfo
	result["siteName"] = siteConfig.Name
	result["SiteUrl"] = siteConfig.Url
	result["path"] = c.Request.URL.Path
	result["refer"] = c.Request.Referer()
	result["version"] = siteConfig.Version
	for _, v := range h {
		for k1, v1 := range v {
			result[k1] = v1
		}
	}
	return result
}

func OutputApi(status int, message string) gin.H {
	result := gin.H{}
	result["status"] = status
	result["message"] = message
	return result
}

// 模板函数：加法运算
func add(a, b int) int {
	return a + b
}

// 模板函数：减法运算
func sub(a, b int) int {
	return a - b
}

// StringInSlice 模板函数：list是否指定字符串运算
func StringInSlice(target string, strList []string) bool {
	for _, str := range strList {
		if str == target {
			return true // 找到了，返回 true
		}
	}
	return false // 循环结束都没找到，返回 false
}

// TruncateString 将字符串截取为指定长度的字符串
func TruncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}
