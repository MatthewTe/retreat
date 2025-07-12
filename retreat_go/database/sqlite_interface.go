package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/list"
	_ "modernc.org/sqlite"
)

type DatabaseArticlesMsg []list.Item
type ArticleItem struct {
	ArticleTitle string
	ParentFeed   string
	PubDate      int64
}

func (i ArticleItem) Title() string       { return i.ArticleTitle }
func (i ArticleItem) Feed() string        { return i.ParentFeed }
func (i ArticleItem) FilterValue() string { return i.ArticleTitle }
func (i ArticleItem) Description() string {
	return fmt.Sprintf("Published: %s | Feed: %s", time.Unix(i.PubDate, 0), i.ParentFeed)
}

func LoadFileFromBlob(localDBPath string) DatabaseArticlesMsg {
	// Given a filepath check the s3 storage bucket for a db. If that db's timestamp is newer than the local version
	// pull it down and return a connection?? path?? to the database
	// Needs to confirm the db is readable and accessable by the applicaton in order to not return an error

	db, err := sql.Open("sqlite", localDBPath)
	if err != nil {
		log.Fatalln(err)
		var articles []list.Item
		return articles
	}

	rows, err := db.Query(`SELECT title, feed, pubDate FROM rss_articles`)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	defer rows.Close()

	var articles []list.Item
	for rows.Next() {
		var article ArticleItem
		if err = rows.Scan(&article.ArticleTitle, &article.ParentFeed, &article.PubDate); err != nil {
			log.Fatalln(err)
			return nil
		}
		articles = append(articles, article)
	}
	if err = rows.Err(); err != nil {
		log.Fatalln(err)
		return articles
	}

	return articles

}

type DatabaseFeedMsg []list.Item
type FeedItem struct {
	FeedTitle   string
	UpdatedDate int64
}

func (i FeedItem) FilterValue() string { return i.FeedTitle }
func (i FeedItem) Title() string       { return i.FeedTitle }
func (i FeedItem) Description() string {
	return fmt.Sprintf("Published: %s", time.Unix(i.UpdatedDate, 0))
}
func LoadFeedsFromDB(localDBPath string) DatabaseFeedMsg {

	db, err := sql.Open("sqlite", localDBPath)
	if err != nil {
		log.Fatalln(err)
		var feeds []list.Item
		return feeds
	}

	rows, err := db.Query(`SELECT title, lastBuildDate FROM rss_feeds`)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	defer rows.Close()

	var feeds []list.Item
	for rows.Next() {
		var feed FeedItem
		if err = rows.Scan(&feed.FeedTitle, &feed.UpdatedDate); err != nil {
			log.Fatalln(err)
			return nil
		}
		feeds = append(feeds, feed)
	}
	if err = rows.Err(); err != nil {
		log.Fatalln(err)
		return feeds
	}

	return feeds
}

func GetArticlesFromFeed(localDBPath string, feedTitle string) DatabaseArticlesMsg {
	db, err := sql.Open("sqlite", localDBPath)
	if err != nil {
		log.Fatalln(err)
		var articles []list.Item
		return articles
	}

	rows, err := db.Query(`SELECT title, feed, pubDate FROM rss_articles WHERE feed = ?`, feedTitle)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	defer rows.Close()

	var articles []list.Item
	for rows.Next() {
		var article ArticleItem
		if err = rows.Scan(&article.ArticleTitle, &article.ParentFeed, &article.PubDate); err != nil {
			log.Fatalln(err)
			return nil
		}
		articles = append(articles, article)
	}
	if err = rows.Err(); err != nil {
		log.Fatalln(err)
		return articles
	}

	return articles

}

func GetArticleMarkdownContent(localDBPath string, articleTitle string) (string, error) {

	db, err := sql.Open("sqlite", localDBPath)
	if err != nil {
		return "", nil
	}

	row := db.QueryRow(`SELECT articleContent FROM rss_articles WHERE title = ?`, articleTitle)

	var articleMarkdown string
	if err = row.Scan(&articleMarkdown); err != nil {
		return "", err
	}

	return articleMarkdown, nil

}
