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

// CreateRenderer 创建一个新的多模板渲染器
func CreateRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	// 创建模板函数映射
	funcMap := map[string]interface{}{
		"add": add,
		"sub": sub,
	}

	// 获取所有模板文件
	base := "templates/base.html"
	templates := []string{
		"templates/home.html",
		"templates/login.html",
		"templates/register.html",
		"templates/link_detail.html",
		"templates/new_link.html",
		"templates/profile.html",
		"templates/user_links.html",
		"templates/tags.html",
		"templates/tag_detail.html",
		"templates/result.html",
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
