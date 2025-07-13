package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charmbracelet/bubbles/list"

	_ "modernc.org/sqlite"
)

func _totalReloadDB(localDBPath string, cfg aws.Config) error {

	s3Client := s3.NewFromConfig(cfg)

	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String("retreat-articles"),
		Key:    aws.String("feed_db.sqlite"),
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()
	file, err := os.Create(localDBPath)
	if err != nil {
		return err
	}
	defer file.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return err
	}

	_, err = file.Write(body)
	return err
}

func SyncDatabsaeFromS3(localDBPath string) error {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	localDBStat, err := os.Stat(localDBPath)
	// Just load if the database does not exist from s3:
	if errors.Is(err, os.ErrNotExist) {
		return _totalReloadDB(localDBPath, cfg)
	}

	// If its some other error throw:
	if err != nil {
		return err
	}

	// File does exists get time from blob:
	client := s3.NewFromConfig(cfg)
	input := &s3.HeadObjectInput{
		Bucket: aws.String("retreat-articles"),
		Key:    aws.String("feed_db.sqlite"),
	}

	s3DBResult, err := client.HeadObject(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to get object metadata: %v", err)
	}

	// Remote database newer than local - download:
	if s3DBResult.LastModified.After(localDBStat.ModTime()) {
		return _totalReloadDB(localDBPath, cfg)
	}

	return err

}

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

func LoadArticlesFromBlob(localDBPath string) DatabaseArticlesMsg {
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
