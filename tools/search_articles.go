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
	"mcp/server/dao"
	"mcp/server/util"
	"sort"
	"strings"
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
	log.Println("🔍 searchArticle called with params:", params)
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

	searchResult, err := client.Qdrant.Query(ctx, q)
	if err != nil {
		log.Println("query qdrant failed, err ", err)
		return mcp.NewToolResultError(fmt.Sprintf("Qdrant search failed: %v", err)), nil
	}

	if len(searchResult) == 0 {
		return mcp.NewToolResultText("未找到相关文章。"), nil
	}

	// 3. 统计命中文章的分布 (Score Map)
	// articleID -> 最高得分
	articleScores := make(map[string]float32)
	// articleID -> 出现的切片列表
	articleChunks := make(map[string][]string)

	for _, hit := range searchResult {
		// 取出 article_id (注意：存入 Qdrant 时必须存这个字段)
		artID := hit.Payload["id"].GetStringValue()
		if artID == "" {
			continue
		}

		// 记录最高分
		if score, exists := articleScores[artID]; !exists || hit.Score > score {
			articleScores[artID] = hit.Score
		}

		// 收集切片文本 (Payload 中的 text 字段)
		chunkText := hit.Payload["textToIndex"].GetStringValue()
		articleChunks[artID] = append(articleChunks[artID], chunkText)
	}

	// 4. 决策策略：我们要读全文还是读切片？
	// 简单策略：如果得分最高的文章 score > 0.85 (非常相关)，且它就是 Top1，那我们就读它的全文
	// 或者：如果 Top 5 里面有 3 个切片都属于同一篇文章，也读全文。

	// 这里我们按得分对文章排序
	var sortedArticles []string
	for id := range articleScores {
		sortedArticles = append(sortedArticles, id)
	}
	sort.Slice(sortedArticles, func(i, j int) bool {
		return articleScores[sortedArticles[i]] > articleScores[sortedArticles[j]]
	})

	topArticleID := sortedArticles[0]
	topScore := articleScores[topArticleID]

	var finalContextBuilder strings.Builder

	// ------------------------------------------------------------------
	// 策略分支 A: 命中非常精准，直接读取长文全文
	// ------------------------------------------------------------------
	if topScore > 0.82 { // 阈值可调，0.82 经验值
		// 调用 DAO 去 MySQL 取 1.3w 字的全文
		fullContent, err := dao.GetFullContentByID(topArticleID)
		if err == nil && fullContent != "" {
			finalContextBuilder.WriteString(fmt.Sprintf("【核心参考文章 (ID:%s)】\n%s\n", topArticleID, fullContent))

			//为了防止漏掉其他关键信息，如果有第二名的文章且分数也不错，可以补充它的摘要
			if len(sortedArticles) > 1 {
				secID := sortedArticles[1]
				if articleScores[secID] > 0.75 {
					sum, _ := dao.GetArticleSummary(secID)
					finalContextBuilder.WriteString(fmt.Sprintf("\n【补充参考】%s\n", sum))
				}
			}

			return mcp.NewToolResultText(fullContent), nil
		}
	}

	// ------------------------------------------------------------------
	// 策略分支 B: 命中比较分散，或者分数不高 -> 组装切片 (RAG 标准模式)
	// ------------------------------------------------------------------
	// 这种情况可能是用户问了一个跨文章的行业问题，比如“新能源车最近有哪些负面？”
	// 我们需要把几个不同文章的切片拼起来。

	for _, artID := range sortedArticles {
		// 简单的去重逻辑
		chunks := articleChunks[artID]
		// 这里的 chunks 只是几百字的小片段
		for _, c := range chunks {
			finalContextBuilder.WriteString(fmt.Sprintf("...%s...\n", c))
		}
		finalContextBuilder.WriteString("\n---\n")
	}

	return mcp.NewToolResultText(finalContextBuilder.String()), nil
}
