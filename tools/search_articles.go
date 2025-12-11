package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/qdrant/go-client/qdrant"
	"log"
	"mcp/server/ai"
	"mcp/server/client"
	"mcp/server/util"
	"strings"
	"time"
)

func getSearchArticleTool() mcp.Tool {
	tool := mcp.NewTool("search_articles",
		mcp.WithDescription(`
Qdrant 语义检索工具：
- 适合：模糊查询、自然语言提问、语义相似度判断
- 不适用：需要按时间排序、获取最新文章、按字段过滤（如 author/date/type）
如果问题涉及 “最新”、“时间排序”、“按字段过滤”、“数据库字段精确筛选”，不要使用本工具，应使用 MySQL 工具。
`),
		mcp.WithString("query", mcp.Description("自然语言问题或长文本查询，将自动生成向量")),
		mcp.WithString("start_time", mcp.Description("开始时间，格式为2006-01-02 15:04:05")),
		mcp.WithString("end_time", mcp.Description("结束时间，格式为2006-01-02 15:04:05")),
		mcp.WithNumber("score", mcp.Description("相似度阈值，浮点数类型，范围0到1，表示返回结果的最低相似度，默认为0.5")),
		mcp.WithNumber("limit", mcp.Description("返回结果数量，默认为 5，最大不超过100")))
	return tool
}

type searchArticleReq struct {
	Query     string  `json:"query"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Limit     int     `json:"limit"`
	Score     float32 `json:"score"`
}

func searchArticle(ctx context.Context, request mcp.CallToolRequest, params string) (*mcp.CallToolResult, error) {
	var searchReq searchArticleReq
	if err := json.Unmarshal([]byte(params), &searchReq); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	queryVec, err := ai.GetEmbedding(ctx, searchReq.Query)
	if err != nil {
		log.Println(fmt.Sprintf("❌ 生成向量失败: %v\n\n", err))
		return mcp.NewToolResultError(fmt.Sprintf("Embedding failed: %v", err)), nil
	}

	q := &qdrant.QueryPoints{
		CollectionName: util.CollectionName,
		Query:          qdrant.NewQuery(queryVec...),
		Limit:          &[]uint64{uint64(searchReq.Limit)}[0],
		WithPayload:    qdrant.NewWithPayload(true),
	}

	if searchReq.StartTime != "" && searchReq.EndTime != "" {
		startTime, _ := time.ParseInLocation(time.DateTime, searchReq.StartTime, util.Loc)
		endTime, _ := time.ParseInLocation(time.DateTime, searchReq.EndTime, util.Loc)
		filter := &qdrant.Filter{
			Should: []*qdrant.Condition{
				qdrant.NewRange("created_at", &qdrant.Range{
					Gte: qdrant.PtrOf(float64(startTime.UTC().Unix())),
					Lte: qdrant.PtrOf(float64(endTime.UTC().Unix())),
				}),
			},
		}
		q.Filter = filter
	}

	searchResult, err := client.Qdrant.Query(ctx, q)
	if err != nil {
		log.Println("query qdrant failed, err ", err)
		return mcp.NewToolResultError(fmt.Sprintf("Qdrant search failed: %v", err)), nil
	}

	// 4. 组装 Prompt (Prompt Engineering)
	var contextBuilder strings.Builder
	for _, point := range searchResult {
		// 只有相似度足够高才用 (阈值过滤)
		if searchReq.Score > 0 && point.Score < searchReq.Score {
			continue
		}
		// 拼接内容
		content := point.Payload["summary"].GetStringValue()
		contextBuilder.WriteString(content)
		contextBuilder.WriteString("\n---\n")
	}

	contextText := contextBuilder.String()
	if contextText == "" {
		contextText = "未找到相关文章。"
	} else {
		log.Printf("📖 找到参考资料 (Top match score: %.4f)\n", searchResult[0].Score)
	}
	return mcp.NewToolResultText(contextText), nil
}
