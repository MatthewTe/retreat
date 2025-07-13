import argparse
import pathlib
import sqlite3
from loguru import logger

from retreat_server.database.setup import init_database
from retreat_server.database.s3 import upload_db_to_s3
from retreat_server.feed_parsers.processing_feeds import check_feeds_for_changes, check_articles_for_changes

if __name__ == "__main__":

    parser = argparse.ArgumentParser()
    parser.add_argument("--database", "-db", help="The sqlite database storing all of the rss feeds", required=True)
    parser.add_argument("--feeds", "-f", nargs="+", default=[], help="Inserts an rss feed into the database for processing")
    parser.add_argument("--load_articles", "-l", action='store_true', help="Syncs all of the feeds and the assocaited articles stored in the database via playwright")
    parser.add_argument("--push_db", "-p", action="store_true", help="Pushing the sqlite database to s3 storage for access by other apps")

    args = parser.parse_args()

    db_path  = pathlib.Path(args.database)
    if not db_path.is_file():
        logger.info(f"{args.database} does not exists - creating database")
        init_database(args.database)
    
    feeds = []
    if len(args.feeds) > 0:
        logger.info(f"Ingesting rss feeds into db:")
        feeds = check_feeds_for_changes(
            feeds=args.feeds,
            db_path=args.database
        )
    
    if args.load_articles:
        logger.info(f"Synching feed articles")
        check_articles_for_changes(feeds, args.database)

    if args.push_db:
        logger.info(f"Pushing database to remote filestore for use by other apps. This will overwrite remote sqlite.db")
        status = upload_db_to_s3(args.database)
        logger.info(f"Database uploaded to blob storage after load {status}")