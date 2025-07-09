from database.get_feeds import get_all_articles, get_all_feeds

if __name__ == "__main__":

    feeds_df = get_all_feeds("./example.db", columns=['title', "url"])
    articles_df = get_all_articles("./example.db", columns=['title', "url", "feed"])
    
    print(feeds_df)