package render

import (
	"html/template"
	"path/filepath"

	"github.com/gin-contrib/multitemplate"
)

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

// CreateRenderer 创建一个新的多模板渲染器
func CreateRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	// 创建模板函数映射
	funcMap := map[string]interface{}{
		"add":            add,
		"sub":            sub,
		"StringInSlice":  StringInSlice,
		"TruncateString": TruncateString,
	}

	// 获取所有模板文件
	base := "templates/base.html"
	templates := []string{
		"templates/home.html",
		"templates/login.html",
		"templates/register.html",
		"templates/link_detail.html",
		"templates/search.html",
		"templates/new_link.html",
		"templates/profile.html",
		"templates/tags.html",
		"templates/tag_detail.html",
		"templates/result.html",
		"templates/articles.html",
		"templates/article_detail.html",
		"templates/new_article.html",
		"templates/update_article.html",
	}

	// 为每个页面配置模板，并应用模板函数
	for _, tmpl := range templates {
		name := filepath.Base(tmpl)
		name = name[:len(name)-len(filepath.Ext(name))]

		// 创建一个新的模板，并应用函数映射
		t := template.New(filepath.Base(base)).Funcs(funcMap)
		t = template.Must(t.ParseFiles(base, tmpl))
		r.Add(name, t)
	}

	return r
}
