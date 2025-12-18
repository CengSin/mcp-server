package dao

import (
	"github.com/PuerkitoBio/goquery"
	"mcp/server/client"
	"strings"
)

func GetFullContentByID(id string) (string, error) {
	var content string
	if err := client.Mysql.Table("article_entries").Select("content").Where("id = ?", id).Scan(&content).Error; err != nil {
		return "", err
	}

	reader, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return "", err
	}
	return reader.Text(), nil
}

func GetArticleSummary(id string) (string, error) {
	var summary string
	if err := client.Mysql.Table("article_entries").Select("content_short").Where("id = ?", id).Scan(&summary).Error; err != nil {
		return "", err
	}

	return summary, nil
}
