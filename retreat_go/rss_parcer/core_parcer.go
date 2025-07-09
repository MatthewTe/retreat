package rssparcer

import (
	"database/sql"
	"log"
	"time"

	"github.com/mmcdole/gofeed"
)

func reconcileFeedTimeStamp(feed *gofeed.Feed) int64 {
	var associatedFeedTimestamp int64
	if feed.UpdatedParsed == nil {

		if feed.PublishedParsed == nil {
			associatedFeedTimestamp = time.Now().Unix()
		} else {
			associatedFeedTimestamp = feed.PublishedParsed.Unix()
		}

	} else {
		associatedFeedTimestamp = feed.UpdatedParsed.Unix()
	}

	return associatedFeedTimestamp

}

func reconcileArticleTimeStamp(item *gofeed.Item) int64 {
	var associatedFeedTimestamp int64
	if item.UpdatedParsed == nil {

		if item.PublishedParsed == nil {
			associatedFeedTimestamp = time.Now().Unix()
		} else {
			associatedFeedTimestamp = item.PublishedParsed.Unix()
		}

	} else {
		associatedFeedTimestamp = item.UpdatedParsed.Unix()
	}

	return associatedFeedTimestamp

}

func CheckFeedsForChanges(feeds []string, db *sql.DB) ([]*gofeed.Feed, error) {

	// Build tuple of (title, unix epoc timestamp):
	var feedsToUpdate []*gofeed.Feed = []*gofeed.Feed{}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	for _, url := range feeds {
		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(url)
		if err != nil {
			return nil, err
		}

		// Reconciling the RSS Feeds dates. Order: UpdateParsed -> PublishedParsed -> Today UTC:
		var associatedFeedTimestamp = reconcileFeedTimeStamp(feed)
		row, execErr := tx.Query(`
			SELECT title 
			FROM rss_feeds
			WHERE (title = ? AND lastBuildDate = ?)
		`,
			feed.Title,
			associatedFeedTimestamp,
		)
		if execErr != nil {
			return nil, execErr
		}

		var titleFromDB string = ""
		for row.Next() {
			row.Scan(&titleFromDB)
		}

		// If the title does not exist then this is a new record:
		if titleFromDB == "" {

			_, execErr = tx.Exec(`
				INSERT INTO rss_feeds (title, url, lastBuildDate, feedContent)
				VALUES (?, ?, ?, ?)
				ON CONFLICT(title) DO UPDATE SET
					url = excluded.url,
					lastBuildDate = excluded.lastBuildDate,
					feedContent = excluded.lastBuildDate;
			`,
				feed.Title,
				feed.Link,
				associatedFeedTimestamp,
				feed.String(),
			)
			if execErr != nil {
				tx.Rollback()
				return nil, execErr
			}

			feedsToUpdate = append(feedsToUpdate, feed)
		}

	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return feedsToUpdate, nil

}

type ArticleContent struct {
	Article          *gofeed.Item
	Content          string
	FeedTitle        string
	ArticleTimestamp int64
}

func CheckArticlesForChanges(feeds []*gofeed.Feed, db *sql.DB) ([]*ArticleContent, error) {

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	// Building the full universe of feed -> items to update:
	var feedNewArticles map[*gofeed.Feed][]*gofeed.Item = make(map[*gofeed.Feed][]*gofeed.Item)
	for _, feed := range feeds {

		var itemsToStore []*gofeed.Item = []*gofeed.Item{}
		for _, article := range feed.Items {

			row, execErr := tx.Query(
				`SELECT title FROM rss_articles WHERE (title = ? AND pubDate = ?)`,
				article.Title,
				reconcileArticleTimeStamp(article),
			)
			if execErr != nil {
				return nil, execErr
			}

			var articleTitle = ""
			for row.Next() {
				row.Scan(&articleTitle)
			}

			// No article so it is flagged to update:
			if articleTitle == "" {
				itemsToStore = append(itemsToStore, article)

			}

		}

		feedNewArticles[feed] = itemsToStore
	}

	// TODO: Make it async
	var processedArticles []*ArticleContent = []*ArticleContent{}
	for rssFeed, feedArticles := range feedNewArticles {
		for _, item := range feedArticles {
			content, err := ExtractContentForFeedArticle(item)
			if err != nil {
				continue
			}
			processedArticles = append(processedArticles, &ArticleContent{
				Article:          item,
				Content:          content,
				FeedTitle:        rssFeed.Title,
				ArticleTimestamp: reconcileArticleTimeStamp(item),
			})
		}
	}

	var updatedArticles []*ArticleContent = []*ArticleContent{}
	for _, processedContent := range processedArticles {

		log.Println("About to insert article:", processedContent.Article.Title)
		_, execErr := db.Exec(`
			INSERT INTO rss_articles (feed, title, url, pubDate, articleContent)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(title) DO UPDATE SET
				feed = excluded.feed,
				url = excluded.url,
				pubDate = excluded.pubDate,
				articleContent = excluded.articleContent;
			`,
			processedContent.FeedTitle,
			processedContent.Article.Title,
			processedContent.Article.Link,
			processedContent.ArticleTimestamp,
			processedContent.Content,
		)
		log.Println("Successfully inserted:", processedContent.Article.Title)
		if execErr != nil {
			return nil, execErr
		}

		updatedArticles = append(updatedArticles, processedContent)
	}

	return updatedArticles, nil
}
