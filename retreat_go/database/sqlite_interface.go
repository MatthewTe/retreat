package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDatabase(db *sql.DB) (*sql.DB, error) {
	// RSS feeds table:
	/*
		Title
		URI to the feed
		lastBuildDate
		Feed content stored as XML blob
	*/

	// RSS article table
	/*
		Parent reference to the [Rss feeds table]
		Title
		URI to the item
		Description (might need an HTML decoder)
		pubDate
		Full markdown content of the article stored as a blob
	*/
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS rss_feeds (
			title TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			lastBuildDate INTEGER,
			feedContent BLOB
		);
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rss_articles (
			feed TEXT NOT NULL,
			title TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			description TEXT,
			pubDate INTEGER,
			articleContent BLOB,
			FOREIGN KEY (feed) REFERENCES rss_feeds(title)
		);
	`)
	if err != nil {
		return nil, err
	}

	//log.Println("Sucessfully created tables rss_feeds and rss_articles in database", db.Stats())

	return db, nil

}
