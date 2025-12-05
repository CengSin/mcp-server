package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/minio/minio-go/v7"
	"mcp/server/client"
	"mcp/server/config"
	"net/url"
	"strings"
	"time"
)

func generateCsvTool() mcp.Tool {
	// 1. 定义工具：generate_document_link
	tool := mcp.NewTool("generate_document_link",
		mcp.WithDescription("将文本内容生成为文件，并返回下载链接。当用户想要保存当前的对话总结、生成的报告或数据表格时使用。"),

		// 参数定义
		mcp.WithString("filename", mcp.Required(), mcp.Description("文件名 (不包含扩展名)，例如 'daily_summary'")),
		mcp.WithString("file_type", mcp.Required(), mcp.Description("文件类型: 'markdown' (用于文本报告), 'csv' (用于表格数据), 'json'")),
		mcp.WithString("content", mcp.Required(), mcp.Description("要保存到文件中的完整文本内容")),
	)
	return tool
}

type GenerateFileReq struct {
	FileName string `json:"filename"`
	FileType string `json:"file_type"`
	Content  string `json:"content"`
}

func generateCsv(ctx context.Context, request mcp.CallToolRequest, params string) (*mcp.CallToolResult, error) {
	var gr GenerateFileReq
	if err := json.Unmarshal([]byte(params), &gr); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// A. 解析参数
	filename := gr.FileName
	fileType := gr.FileType
	content := gr.Content

	if filename == "" || content == "" {
		return mcp.NewToolResultError("Filename and content are required"), nil
	}

	// B. 处理后缀和 MIME type
	var ext, mimeType string
	switch fileType {
	case "csv":
		ext = ".csv"
		mimeType = "text/csv"
	case "json":
		ext = ".json"
		mimeType = "application/json"
	default:
		ext = ".md"
		mimeType = "text/markdown"
	}

	// C. 构建完整文件名
	// 加上时间戳防止重名: market_report_170123456.md
	fullObjectName := fmt.Sprintf("generated/%s_%d%s", filename, time.Now().Unix(), ext)

	// D. 上传到 MinIO (调用上一轮封装好的 Upload 方法)
	// 注意：这里我们直接把 content 字符串转为 byte 数组上传，不需要存本地文件
	preUrl, err := UploadStringContentToMinIO(ctx, fullObjectName, content, mimeType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Upload failed: %v", err)), nil
	}

	// E. 返回给 LLM
	return mcp.NewToolResultText(fmt.Sprintf("文件已生成。下载链接: %s", preUrl)), nil

}

// 辅助函数：直接上传字符串内容
func UploadStringContentToMinIO(ctx context.Context, objectName, content, contentType string) (string, error) {
	BucketName := config.Cfg.MinIO.BucketName
	reader := strings.NewReader(content)
	// 3. 上传到 MinIO
	// 使用 PutObject 直接上传内存流，不需要存本地磁盘
	_, err := client.MinIO.PutObject(ctx, BucketName, objectName, reader, int64(reader.Len()), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// 4. 生成预签名下载链接 (Presigned URL)
	// 有效期设为 1 小时
	expiry := time.Hour * 1

	// 设置响应头：让浏览器强制下载，而不是在页面里打开
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectName))

	presignedURL, err := client.MinIO.PresignedGetObject(ctx, BucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}

	// 参考上一轮的代码实现
	return presignedURL.String(), nil
}
