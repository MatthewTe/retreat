from playwright.sync_api import sync_playwright
from retreat_server.feed_parsers.article_content_extraction import extract_content_via_playwright

from urllib.parse import urlparse, parse_qs

from newspaper import Article, fulltext

if __name__ == "__main__":

    url = "https://www.google.com/url?rct=j&sa=t&url=https://en.news1.kr/northkorea/5684800&ct=ga&cd=CAIyGzdjMzUzYjRjODYyMzBmODA6Y2E6ZW46S1I6Ug&usg=AOvVaw0uvVQKcm3fcGTJiQUrtMgt"
    
    parsed_url = urlparse(url)

    # Catch query for google links to extract raw url from google alerts rss feed:
    if "google" in parsed_url.hostname and "url" in parsed_url.query:
        extracted_url = parse_qs(parsed_url.query)['url'][0]
    else:
        extracted_url = url

    with sync_playwright() as p:
        browser = p.firefox.launch(headless=True)
        page = browser.new_page()

        try:
            reader_mode_url = f"about:reader?url={extracted_url}"
            page.goto(reader_mode_url, wait_until="domcontentloaded")
            page.wait_for_selector("article")

            content = page.locator("article").inner_text()
        except:
            page.goto(extracted_url, wait_until="domcontentloaded")
            page.wait_for_selector("body"),
            html_content =  page.locator("body").inner_html()
            content = fulltext(html_content)

        #content = extract_content_via_playwright(p, "https://www.google.com/url?rct=j&sa=t&url=https://en.news1.kr/northkorea/5684800&ct=ga&cd=CAIyGzdjMzUzYjRjODYyMzBmODA6Y2E6ZW46S1I6Ug&usg=AOvVaw0uvVQKcm3fcGTJiQUrtMgt")

    print(content)