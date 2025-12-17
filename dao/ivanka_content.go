package dao

import (
	"github.com/PuerkitoBio/goquery"
	"os"
)

func GetFullContentByID(id string) (string, error) {
	file, err := os.Open("test.html")
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return "", err
	}
	return reader.Text(), nil
}
