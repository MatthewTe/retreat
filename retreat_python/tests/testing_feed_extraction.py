from feed_parsers.processing_feeds import check_feeds_for_changes, check_articles_for_changes
from database.setup import init_database

if __name__ == "__main__":
    
    rss_feeds = [
        "https://www.google.ca/alerts/feeds/04364436683859664111/1542882698261322396"
    ]   
    
    init_database("./example.db")

    feeds = check_feeds_for_changes(
        rss_feeds,
        "./example.db"
    )

    articles = check_articles_for_changes(feeds, "./example.db")

    import pprint
    pprint.pprint(articles)