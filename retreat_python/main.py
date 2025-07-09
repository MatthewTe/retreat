import os
import sys

from textual.app import App, ComposeResult
from textual.widgets import DataTable, Footer, ContentSwitcher, Markdown

from textual import events

from rich import print
from rich.columns import Columns
from rich.text import Text

from database.get_feeds import get_all_feeds, get_all_articles, get_article_content

DATABASE = "/Users/matthewteelucksingh/Repos/retreat/retreat_server/.example.db"

class MainApp(App):

    def compose(self) -> ComposeResult:

        self.table = DataTable(id="feed_datatable")
        self.article_content = Markdown(id="article_markdown", markdown=None)
        self.footer = Footer()
        
        
        with ContentSwitcher(initial="feed_datatable"):
            yield self.table
            yield self.article_content

        yield self.footer

    def on_mount(self) -> None:
        self.table.add_columns("title", "pubDate", "feed")
        feeds_df = get_all_articles(DATABASE, columns=['title', 'pubDate', "feed"])
        self.table.cursor_type = "row"

        for row in feeds_df.itertuples(index=False):
            self.table.add_row(row.title, row.pubDate, row.feed, key=row.title)

    # Tracking row state for search query:
    def on_data_table_row_highlighted(self, event: DataTable.RowSelected) -> None:
        self.selected_article_key = event.row_key.value

    def on_key(self, event: events.Key) -> None:
        if event.key == "q":
            self.exit()
        
        match self.query_one(ContentSwitcher).current:
            case "feed_datatable":
                
                match event.key: 
                    case "j":
                        self.table.cursor_coordinate = self.table.cursor_coordinate.down()

                    case "k":
                        self.table.cursor_coordinate = self.table.cursor_coordinate.up()

                    case "enter":
                        content_from_db = get_article_content(self.selected_article_key, DATABASE)
                        self.article_content.update(content_from_db)
                        self.query_one(ContentSwitcher).current = "article_markdown"

            case "article_markdown":
                match event.key:
                    case "b":
                      self.query_one(ContentSwitcher).current = "feed_datatable"

if __name__ == "__main__":

    app = MainApp()
    app.run(inline=True)