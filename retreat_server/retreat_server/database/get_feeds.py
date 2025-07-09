import pandas as pd
import sqlite3
from typing import Optional, List, Dict, Any

def get_all_feeds(db_path: str, columns: list[str] = None) -> pd.DataFrame:
    conn = sqlite3.connect(db_path)

    if columns:
        column_str = ", ".join(columns)
    else:
        column_str = "title, url, lastBuildDate, feedContent"

    query = f"""
        SELECT {column_str}
        FROM rss_feeds
    """

    df = pd.read_sql_query(query, conn)
    conn.close()
    return df

def get_all_articles(
    db_path: str, 
    columns: Optional[List[str]] = None,
    filters: Optional[Dict[str, Any]] = None
) -> pd.DataFrame:
    conn = sqlite3.connect(db_path)

    if columns:
        column_str = ", ".join(columns)
    else:
        column_str = "feed, title, url, pubDate, articleContent"

    where_clause = ""
    params = []
    if filters:
        conditions = [f"{k} = ?" for k in filters.keys()]
        where_clause = "WHERE " + " AND ".join(conditions)
        params = list(filters.values())

    query = f"""
        SELECT {column_str}
        FROM rss_articles
        {where_clause}
    """

    df = pd.read_sql_query(query, conn, params=params)
    conn.close()
    return df

def get_article_content(title: str, db_path: str) -> str:

    conn = sqlite3.connect(db_path)

    cur = conn.execute("SELECT articleContent FROM rss_articles WHERE title = ?", (title,))
    full_content = cur.fetchone()
    conn.close()

    return full_content