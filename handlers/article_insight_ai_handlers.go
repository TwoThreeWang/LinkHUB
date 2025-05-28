package handlers

import (
	"LinkHUB/config"
	"LinkHUB/models"
	"LinkHUB/utils"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)
type SummaryRequest struct {
	CONTENT  string `json:"content" binding:"required"`
	TYPE string `json:"type" binding:"required"`
	CfTurnstile string `json:"cf_turnstile"`
}

// HandleSummarize 处理文章总结请求
func HandleSummarize(c *gin.Context) {
	// 从上下文中获取用户信息
	userInfo := GetCurrentUser(c)
	if userInfo == nil {
		utils.RespFail(c, http.StatusForbidden, "仅限登陆后使用")
		return
	}
	// 从上下文获取配置
	cfg := config.GetConfig()
	if cfg == nil {
		utils.RespFail(c, http.StatusBadRequest, "配置未找到")
		return
	}

	// 初始化服务
	aiService, err := NewAIService(cfg)
	if err != nil {
		utils.RespFail(c, http.StatusBadRequest, "AI 服务初始化报错：" + err.Error())
		return
	}

	// 获取请求参数
	var req SummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespFail(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// 验证 Turnstile 令牌
	if req.CfTurnstile != "" {
		remoteIP := c.ClientIP()
		_, err := utils.VerifyTurnstileToken(c, req.CfTurnstile, remoteIP)
		if err!= nil {
			utils.RespFail(c, http.StatusBadRequest, "验证 Turnstile 令牌失败：" + err.Error())
			return
		}
	}else{
		utils.RespFail(c, http.StatusBadRequest, "验证 Turnstile 令牌失败：" + "缺少验证参数")
		return
	}

	// 获取页面内容
	var content string
	if req.TYPE == "url" {
		fetchService := NewFetchService()
		content, err = fetchService.FetchStory(req.CONTENT)
		if err != nil {
			utils.RespFail(c, http.StatusBadRequest, "获取文章内容失败：" + err.Error())
			return
		}
	} else {
		content = fmt.Sprintf("\n<article>\n%s\n</article>\n", req.CONTENT)
	}

	story := &models.Story{Content: content}

	// AI 文章总结
	err = aiService.GenerateSummary(story)
	if err != nil {
		utils.RespFail(c, http.StatusBadRequest, "总结文章内容失败：" + err.Error())
		return
	}
	// 返回数据
	utils.RespSuccess(c, gin.H{
		"summary":  story.Summary,
	})
}


// 获取文章相关函数
type FetchService struct {
	client *http.Client
}

func NewFetchService() *FetchService {
	return &FetchService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchStory 获取单个文章的详细信息
func (s *FetchService) FetchStory(url string) (string, error) {
	// 设置请求头
	headers := make(http.Header)
	//headers.Set("X-Retain-Images", "none")

	// 获取文章内容
	req, err := http.NewRequest("GET", "https://r.jina.ai/"+url, nil)
	if err != nil {
		return "", err
	}
	req.Header = headers

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s %s", resp.Status, url)
	}

	articleBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 构建返回内容
	content := fmt.Sprintf("\n<article>\n%s\n</article>\n", string(articleBody))

	return content, nil
}

// 请求 AI 相关函数
type AIService struct {
	config  *config.Config
	client  *http.Client
	service *genai.Client
}

func NewAIService(cfg *config.Config) (*AIService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("初始化Gemini客户端失败: %v", err)
	}

	return &AIService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		service: client,
	}, nil
}

// GenerateSummary 为文章生成中文总结
func (s *AIService) GenerateSummary(story *models.Story) error {
	// 构建提示词
	prompt := fmt.Sprintf(
		`你是一位中文博客的编辑助理，擅长将文章整理成引人入胜的博客内容。内容受众主要为软件开发者和科技爱好者。

【工作目标】
- 接收并阅读文章。
- 为文章起一个有利于 SEO，并且能够吸引人点击的标题，标题不要太长，并且标题要贴合文章内容。
- 先简明介绍文章的主要话题，再对其要点进行精炼说明。
- 以清晰直接的口吻进行讨论，像与朋友交谈般简洁易懂。
- 按照逻辑顺序，使用二级标题 (如"## 标题") 与分段正文形式呈现播客的核心精简内容。
- 所有违反中国大陆法律和政治立场的内容，都跳过。

【输出要求】
- 直接输出正文，不要返回前言。
- 直接进入主要内容的总结与讨论：
  * 第 1 句使用使用一级标题输出文章生成的文章标题。
  * 第 2-3 句：概括适合搜索引擎收录的文章主题。
  * 第 4-15 句：详细阐述文章的重点内容。
  * 第 16-20 句：总结分析，体现多角度探讨。
- 直接返回 Markdown 格式的正文内容。
- 换行不要使用\n,使用两个回车。
- 必须使用简体中文输出。`,
		story.Content,
	)

	// 调用Gemini API生成总结
	ctx := context.Background()
	model := s.service.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.3)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("生成总结失败: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return fmt.Errorf("未能生成有效的总结")
	}

	// 更新文章的总结信息
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		content := resp.Candidates[0].Content.Parts[0]
		story.Summary = fmt.Sprintf("%s", content)
		return nil
	} else {
		return fmt.Errorf("未能获取到有效的总结内容")
	}
}