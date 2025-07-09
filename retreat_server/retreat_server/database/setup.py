import sqlite3
from typing import Optional, Tuple

def init_database(db_path: str) -> Tuple[Optional[sqlite3.Connection], Optional[Exception]]:
    """
    Initialize the SQLite database with rss_feeds and rss_articles tables.
    
    :param db_path: Path to the SQLite database file.
    :return: Tuple of (sqlite3.Connection or None, Exception or None)
    """
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()

        # Create rss_feeds table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS rss_feeds (
                title TEXT PRIMARY KEY,
                url TEXT NOT NULL,
                lastBuildDate INTEGER,
                feedContent BLOB
            );
        """)

        # Create rss_articles table
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS rss_articles (
                feed TEXT NOT NULL,
                title TEXT PRIMARY KEY,
                url TEXT NOT NULL,
                description TEXT,
                pubDate INTEGER,
                articleContent BLOB,
                FOREIGN KEY (feed) REFERENCES rss_feeds(title)
            );
        """)

        conn.commit()
        return conn, None

    except Exception as e:
        return None, e
