package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// AIService 封装 Eino 的 AI 生成逻辑
type AIService struct {
	chatModel model.ChatModel
	apiKey    string
}

// NewAIService 现在接收更多的配置项，从而支持多种 AI 模型供应商
func NewAIService(apiKey, baseURL, modelName string) *AIService {
	if apiKey == "" {
		log.Println("WARNING: AI_API_KEY is empty, AI generation will fail.")
	}
	
	m, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,   // 从环境变量加载
		Model:   modelName, // 从环境变量加载
	})
	if err != nil {
		log.Fatalf("failed to init chat model: %v", err)
	}

	return &AIService{
		chatModel: m,
		apiKey:    apiKey,
	}
}

func (s *AIService) Generate(ctx context.Context, input string) (*Blog, error) {
	// 1. 生成博文文本内容
	resp, err := s.chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage("你是一个专业的博客作者。请根据输入的主题生成一篇博客。格式要求：第一行是标题(Title:XXX)，剩下的内容是正文。"),
		schema.UserMessage(input),
	})
	if err != nil {
		return nil, err
	}

	fullContent := resp.Content
	title := "AI 创作: " + input
	content := fullContent

	// 解析 AI 输出
	if strings.Contains(fullContent, "Title:") {
		parts := strings.SplitN(fullContent, "\n", 2)
		title = strings.TrimPrefix(parts[0], "Title: ")
		if len(parts) > 1 {
			content = parts[1]
		}
	}

	// 2. AI 图像创建逻辑：基于 pollinations 免费 API 提供真正的图像输出
	// imageURL 使用用户的 Prompt 来生成具有高度关联性的博客封面图
	imagePrompt := strings.ReplaceAll(input, " ", ",")
	imageURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s-vivid-blog-cover-cinematic-lighting?width=1080&height=720&nologo=true", imagePrompt)

	return &Blog{
		Title:    title,
		Content:  content,
		ImageURL: imageURL,
	}, nil
}
