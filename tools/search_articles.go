package tools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/qdrant/go-client/qdrant"
	"log"
	"mcp/server/ai"
	"mcp/server/client"
	"mcp/server/util"
	"strings"
)

func getSearchArticleTool() mcp.Tool {
	tool := mcp.NewTool("search_articles",
		mcp.WithDescription("根据自然语言查询金融文章数据库。支持语义搜索。"),
		mcp.WithString("query", mcp.Required(), mcp.Description("用户的搜索关键词或问题")),
		mcp.WithNumber("limit", mcp.Description("返回结果数量，默认为 5")))
	return tool
}

func searchArticle(ctx context.Context, request mcp.CallToolRequest, question string) (*mcp.CallToolResult, error) {
	if len(question) == 0 {
		return mcp.NewToolResultError("Query argument is required"), nil
	}
	queryVec, err := ai.GetEmbedding(ctx, question)
	if err != nil {
		log.Println(fmt.Sprintf("❌ 生成向量失败: %v\n\n", err))
		return mcp.NewToolResultError(fmt.Sprintf("Embedding failed: %v", err)), nil
	}

	searchResult, err := client.Qdrant.Query(ctx, &qdrant.QueryPoints{
		CollectionName: util.CollectionName,
		Query:          qdrant.NewQuery(queryVec...),
		Limit:          &[]uint64{3}[0],
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		log.Println("query qdrant failed, err ", err)
		return mcp.NewToolResultError(fmt.Sprintf("Qdrant search failed: %v", err)), nil
	}

	// 4. 组装 Prompt (Prompt Engineering)
	var contextBuilder strings.Builder
	for _, point := range searchResult {
		// 只有相似度足够高才用 (阈值过滤)
		if point.Score > 0.5 {
			content := point.Payload["summary"].GetStringValue()
			contextBuilder.WriteString(content)
			contextBuilder.WriteString("\n---\n")
		}
	}

	contextText := contextBuilder.String()
	if contextText == "" {
		contextText = "未找到相关文章。"
	} else {
		log.Printf("📖 找到参考资料 (Top match score: %.4f)\n", searchResult[0].Score)
	}
	return mcp.NewToolResultText(contextText), nil
}
