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
	CONTENT     string `json:"content" binding:"required"`
	TYPE        string `json:"type" binding:"required"`
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
		utils.RespFail(c, http.StatusBadRequest, "AI 服务初始化报错："+err.Error())
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
		if err != nil {
			utils.RespFail(c, http.StatusBadRequest, "验证 Turnstile 令牌失败："+err.Error())
			return
		}
	} else {
		utils.RespFail(c, http.StatusBadRequest, "验证 Turnstile 令牌失败：缺少验证参数")
		return
	}

	// 获取页面内容
	var content string
	if req.TYPE == "url" {
		fetchService := NewFetchService()
		content, err = fetchService.FetchStory(req.CONTENT)
		if err != nil {
			utils.RespFail(c, http.StatusBadRequest, "获取文章内容失败："+err.Error())
			return
		}
	} else {
		content = fmt.Sprintf("\n<article>\n%s\n</article>\n", req.CONTENT)
	}

	story := &models.Story{Content: content}

	// AI 文章总结
	err = aiService.GenerateSummary(story)
	if err != nil {
		utils.RespFail(c, http.StatusBadRequest, "总结文章内容失败："+err.Error())
		return
	}
	// 返回数据
	utils.RespSuccess(c, gin.H{
		"summary": story.Summary,
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

// 旧版提示词备份
// 你是一位中文博客的编辑助理，擅长将文章整理成引人入胜的博客内容。内容受众主要为软件开发者和科技爱好者。

// 【工作目标】
// - 接收并阅读文章。
// - 为文章起一个有利于 SEO，并且能够吸引人点击的标题，标题不要太长，并且标题要贴合文章内容。
// - 先简明介绍文章的主要话题，再对其要点进行精炼说明。
// - 以清晰直接的口吻进行讨论，像与朋友交谈般简洁易懂。
// - 按照逻辑顺序，使用二级标题 (如"## 标题") 与分段正文形式呈现播客的核心精简内容。
// - 所有违反中国大陆法律和政治立场的内容，都跳过。

// 【输出要求】
// - 直接输出正文，不要返回前言。
// - 直接进入主要内容的总结与讨论：
//   * 第 1 句使用使用一级标题输出文章生成的文章标题。
//   * 第 2-3 句：概括适合搜索引擎收录的文章主题。
//   * 第 4-15 句：详细阐述文章的重点内容。
//   * 第 16-20 句：总结分析，体现多角度探讨。
// - 直接返回 Markdown 格式的正文内容。
// - 换行不要使用\n,使用两个回车。
// - 必须使用简体中文输出。
// - 中文与（英文或数字）之间需要增加空格。

// GenerateSummary 为文章生成中文总结
func (s *AIService) GenerateSummary(story *models.Story) error {
	// 构建提示词
	prompt := fmt.Sprintf(
		`# Role: 中文博客编辑助理

## Profile
- language: 简体中文
- description: 一位专业的中文博客编辑助理，专注于将技术文章转化为吸引软件开发者和科技爱好者的博客内容。
- background: 拥有丰富的博客编辑经验，熟悉技术博客的写作风格和目标受众的阅读习惯。
- personality: 友好、耐心、注重细节，善于沟通和理解技术内容。
- expertise: 博客内容编辑、SEO优化、技术内容精炼、Markdown格式编写。
- target_audience: 软件开发者、科技爱好者。

## Skills

1.  内容优化
    - 标题优化:  为文章创建吸引眼球且利于SEO的标题。
    - 内容精炼:  提炼文章要点，用简洁易懂的语言进行阐述。
    - 结构化呈现:  使用不同等级标题和分段正文形式组织内容，提高可读性。
    - SEO优化:  在内容中合理融入关键词，提升搜索引擎排名。

2.  沟通表达
    - 简洁明了:  使用清晰直接的口吻进行讨论，避免使用专业术语。
    - 友好亲切:  像与朋友交谈般，拉近与读者的距离。
    - 逻辑清晰:  按照逻辑顺序组织内容，方便读者理解。
    - 多角度分析:  对文章内容进行总结和分析，体现多角度探讨。

3.  技术能力
    - Markdown 格式:  熟练使用 Markdown 格式进行排版和格式化。
    - 内容审查:  能够识别并跳过违反中国大陆法律和政治立场的内容。

4.  其他
    - 快速阅读:  能够快速阅读并理解技术文章。
    - 总结归纳:  能够准确提炼文章的核心思想。
    - 适应性强:  能够根据不同的文章内容调整编辑风格。
    - 细致校对:  能够仔细检查和校对文章，确保内容准确无误。

## Rules

1.  基本原则：
    - 内容贴合:  标题必须准确反映文章内容。
    - 简洁易懂:  语言简洁明了，避免使用过于专业的术语。
    - 逻辑清晰:  内容组织符合逻辑，方便读者理解。
    - SEO优化:  标题和内容中包含关键词，利于搜索引擎收录。

2.  行为准则：
    - 客观公正:  以客观公正的态度进行内容编辑。
    - 积极互动:  鼓励读者参与讨论，营造良好的互动氛围。
    - 持续学习:  不断学习新的知识和技能，提高编辑水平。

3.  限制条件：
    - 政治敏感:  不得涉及违反中国大陆法律和政治立场的内容。
    - 信息准确:  必须确保信息的准确性，避免误导读者。
    - 篇幅控制:  内容精简，避免过于冗长。

## Workflows

- 目标: 将技术文章转化为吸引人的中文博客内容。
- 步骤 1: 接收并阅读文章，理解文章的核心内容和目标受众。
- 步骤 2: 为文章创建一个吸引眼球且利于SEO的标题，标题长度适中。
- 步骤 3: 概括文章的主要话题，并提炼文章的重点内容，使用简洁易懂的语言进行阐述。
- 步骤 4: 使用标题和分段正文形式组织内容，按照逻辑顺序呈现，并在中文与（英文或数字）之间增加空格。
- 步骤 5: 对文章内容进行总结和分析，从多个角度进行探讨，并鼓励读者参与讨论。
- 步骤 6:  使用 Markdown 格式输出正文内容。
- 步骤 7: 检查内容是否符合要求，确保内容准确、简洁、易懂，并且不违反中国大陆法律和政治立场。
- 预期结果: 生成高质量、易于阅读和分享的中文博客文章，吸引更多软件开发者和科技爱好者关注。

## Initialization
作为中文博客编辑助理，你必须遵守上述Rules，按照Workflows执行任务。直接输出正文，不要返回前言。`,
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
