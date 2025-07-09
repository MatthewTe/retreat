from playwright.sync_api import Page

from newspaper import Article, fulltext
from urllib.parse import urlparse, parse_qs
from markdownify import markdownify
from loguru import logger
def extract_content_via_playwright(page: Page, url: str) -> str:
    
    parsed_url = urlparse(url)

    # Catch query for google links to extract raw url from google alerts rss feed:
    if "google" in parsed_url.hostname and "url" in parsed_url.query:
        extracted_url = parse_qs(parsed_url.query)['url'][0]
    else:
        extracted_url = url

    try:
        logger.info(f"Attempting to extract info from: {extracted_url}")
        reader_mode_url = f"about:reader?url={extracted_url}"
        page.goto(reader_mode_url, wait_until="domcontentloaded")
        page.wait_for_timeout(1500)
        page.wait_for_selector("body", timeout=100)
        html_content = page.locator("body").inner_html()
        content = markdownify(html_content)
    except:
        page.goto(extracted_url, wait_until="domcontentloaded")
        page.wait_for_timeout(1500)
        html_content = page.locator("body").inner_html()
        page.wait_for_selector("body", timeout=100),
        html_content =  page.locator("body").inner_html()
        try:
            content = markdownify(html_content)
        except:
            content = ""

    return content