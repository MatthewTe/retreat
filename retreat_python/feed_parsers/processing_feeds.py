import sqlite3
import time
from typing import List, Optional, Tuple, Dict, Any, TypedDict
from loguru import logger
import pprint
from bs4 import BeautifulSoup
import traceback

from rss_parser import RSSParser, AtomParser
from rss_parser.models.rss import RSS
from rss_parser.models import XMLBaseModel

import feedparser
import curl_cffi
from urllib.parse import urlparse, parse_qs

from readability import Document
from newspaper import Article, fulltext

import requests
from time import mktime
from datetime import datetime

def reconcile_article_timestamp(item: dict) -> int:
    updated = item.get("updated")
    published = item.get("published")

    if updated:
        try:
            return int(datetime.fromisoformat(updated).timestamp())
        except Exception:
            pass

    if published:
        try:
            return int(datetime.fromisoformat(published).timestamp())
        except Exception:
            pass

    return int(time.time())

class Feed(TypedDict):
    title: str
    updated: str 
    updated_parsed: str 
   
class CoreRSSFeeds(TypedDict):
    entries: list[dict]
    feed: Feed

def reconcile_feed_timestamp(feed_metadata: Feed) -> int:
    updated = feed_metadata.get("updated_parsed")
    published = feed_metadata.get("published_parsed")

    if updated:
        try:
            return int(mktime(updated))
        except Exception:
            pass

    if published:
        try:
            return int(mktime(published))
        except Exception:
            pass

    return int(time.time())

def check_feeds_for_changes(feeds: List[str], db_path: str) -> list[CoreRSSFeeds]:
    feeds_to_update = []

    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    for url in feeds:
        response = requests.get(url)
        response.raise_for_status()
        
        feed: CoreRSSFeeds = feedparser.parse(response.text)
        feed_core_metadata: Feed = feed['feed']
        associated_timestamp = reconcile_feed_timestamp(feed_core_metadata)

        title = feed["feed"].get("title", "")

        cursor.execute("""
            SELECT title FROM rss_feeds
            WHERE title = ? AND lastBuildDate = ?
        """, (title, associated_timestamp))
        row = cursor.fetchone()

        if not row:
            # Feed is new or updated
            cursor.execute("""
                INSERT INTO rss_feeds (title, url, lastBuildDate, feedContent)
                VALUES (?, ?, ?, ?)
                ON CONFLICT(title) DO UPDATE SET
                    url = excluded.url,
                    lastBuildDate = excluded.lastBuildDate,
                    feedContent = excluded.lastBuildDate
            """, (
                title,
                url,
                associated_timestamp,
                response.text
            ))
            feeds_to_update.append(feed)

    conn.commit()
    return feeds_to_update

def reconcile_article_timestamp(item: dict) -> int:
    updated = item.get("updated")
    published = item.get("published")

    if updated:
        try:
            return int(datetime.fromisoformat(updated).timestamp())
        except Exception:
            pass

    if published:
        try:
            return int(datetime.fromisoformat(published).timestamp())
        except Exception:
            pass

    return int(time.time())

class CoreRSSItem(TypedDict):
    link: str
    published_parsed: Tuple
    updated_parsed: Tuple
    title: str

def extract_content_for_feed_article(item: CoreRSSItem) -> Optional[str]:

    try:
        parsed_url = urlparse(item['link'])

        # Catch query for google links to extract raw url from google alerts rss feed:
        if "google" in parsed_url.hostname and "url" in parsed_url.query:
            extracted_url = parse_qs(parsed_url.query)['url'][0]
        else:
            extracted_url = item['link']
        
        response = curl_cffi.get(extracted_url, impersonate="chrome110")

        print(response.text)
        import sys
        sys.exit(0)

        soup = BeautifulSoup(response.text)
        body_content = soup.find("body")
        text =  fulltext(body_content)

        print(text)

        return text

    except Exception as e:
        logger.error(f"Error extracting content from item {e}")
        return None

def check_articles_for_changes(feeds: List[CoreRSSFeeds], db_path: str) -> Tuple[List[Dict[str, Any]], Optional[Exception]]:
    updated_articles = []

    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("PRAGMA foreign_keys = ON")
    conn.isolation_level = "DEFERRED"  # Begin transaction

    feed_new_articles: dict[str, list[CoreRSSItem]] = {}

    for feed in feeds:
        feed_title = feed["feed"]["title"]
        new_items: list[CoreRSSItem] = []

        for item in feed.get("entries", []):
            timestamp = reconcile_feed_timestamp(item)

            cursor.execute("""
                SELECT title FROM rss_articles
                WHERE title = ? AND pubDate = ?
            """, (item.get("title", ""), timestamp))
            row = cursor.fetchone()

            if not row:
                new_items.append(item)

        feed_new_articles[feed_title] = new_items

    # Process & insert new articles
    for feed_title, articles in feed_new_articles.items():

        for item in articles:
            content = extract_content_for_feed_article(item)
            timestamp = reconcile_feed_timestamp(item)

            cursor.execute("""
                INSERT INTO rss_articles (feed, title, url, pubDate, articleContent)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(title) DO UPDATE SET
                    feed = excluded.feed,
                    url = excluded.url,
                    pubDate = excluded.pubDate,
                    articleContent = excluded.articleContent
            """, (
                feed_title,
                item.get("title", ""),
                item.get("link", ""),
                timestamp,
                content
            ))

            updated_articles.append({
                "feed": feed_title,
                "title": item.get("title", ""),
                "url": item.get("link", ""),
                "timestamp": timestamp,
                "content": content
            })

    conn.commit()
    conn.close()
    return updated_articles