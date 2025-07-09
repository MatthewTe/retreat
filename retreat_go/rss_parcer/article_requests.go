package rssparcer

import (
	"net/http"

	"github.com/cixtor/readability"
	"github.com/mmcdole/gofeed"
)

func ExtractContentForFeedArticle(article *gofeed.Item) (string, error) {

	return "HELLO WORLD", nil

	resp, err := http.NewRequest(http.MethodGet, article.Link, nil)
	if err != nil {
		return "", err
	}

	reader := readability.New()
	parsedArticle, err := reader.Parse(resp.Body, article.Link)
	if err != nil {
		return "", err
	}

	return parsedArticle.TextContent, nil
}
