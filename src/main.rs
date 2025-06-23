use std::{collections::HashMap};
use dom_smoothie::{Article, Config, Readability, TextMode};

use rss::Channel;
use color_eyre::{eyre::{Error, Ok}, Result};
use crossterm::{event::{self, Event, KeyCode, KeyEvent, KeyEventKind, KeyModifiers}};
use ratatui::{
    layout::{self, Constraint, Direction}, text::{Line, Span}, widgets::{Block, BorderType, Borders, Cell, List, ListItem, ListState, Paragraph, Row, Scrollbar, ScrollbarState, Table, TableState, Wrap}, DefaultTerminal, Frame
};
use clap::Parser;

#[derive(Parser, Debug)]
struct Args {

  #[arg(default_value="")]
  feeds: String,
}

fn main() -> Result<(), Error>{
    
    let args = Args::parse();

    color_eyre::install()?;
    let terminal = ratatui::init();
    
    let mut app: App = App::new();
    let rss_feeds = extract_rss_feed_lst_from_url(args.feeds).unwrap();
    let _ = app.load_rss_feeds(rss_feeds);

    let result = app.run(terminal);
    ratatui::restore();
    result

}


#[derive(Debug, Default)]
enum AppState {
    #[default]
    FeedView,
    ArticleView,
    ReadingView
}

#[derive(Debug, Default)]
struct RssFeedList {
    items: Vec<Channel>,
    state: ListState,
}

// TODO: Add logic to parse cli

/// The main application which holds the state and logic of the application.
#[derive(Debug, Default)]
pub struct App {

    /// Is the application running?
    running: bool,

    // The list of RSS feeds connected (loaded externally)
    rss_feeds: RssFeedList,

    // The selected Rss Feed:
    selected_rss_feed_tbl_state: TableState,

    existing_articles_map: HashMap<String, String>,
    rss_feed_content_map: HashMap<String, Channel>,

    // Vertical Scroll:
    vertical_scroll: usize,
    vertical_scroll_state: ScrollbarState,

    current_view_state: AppState

}

fn extract_rss_feed_lst_from_url(url: String) -> Result<Vec<String>, Error> {

    let raw_page_content = reqwest::blocking::get(url)?
        .text()?;

    let all_feeds: Vec<String> = raw_page_content
        .split(",")
        .map(|s| s.to_string())
        .collect();
    
    Ok(all_feeds)
   
}

fn extract_full_content_from_url(url: String) -> Result<String, Error> {
    let raw_page_content = reqwest::blocking::get(url)?
        .text()?;

    // let page_document = Document::from(raw_page_content);
    
    // for more options check the documentation
    let cfg = Config {
        text_mode: TextMode::Formatted,
        max_elements_to_parse: 9000,
        ..Default::default()
    };


    let mut readability = Readability::new(
        raw_page_content,
        None,
        Some(cfg))?;

    let article: Article = readability.parse()?;

    Ok(article.text_content.to_string())
}

fn extract_rss_feed_channel_from_url(url: String) -> Result<Channel, Error> {

    let feed_content = reqwest::blocking::get(url)?
        .bytes()?;

    let rss_channel = Channel::read_from(&feed_content[..]);
    
    Ok(rss_channel?)

}

impl App {
    /// Construct a new instance of [`App`].
    pub fn new() -> Self {
        Self::default()
    }

    // Load the Rss Feeds directly:
    pub fn load_rss_feeds(&mut self, rss_feeds: Vec<String>) -> Result<(), Error> {

        let rss_feed_iter = rss_feeds.iter();
        let mut loaded_rss_feed: Vec<Channel> = Vec::new();

        for feed_url in rss_feed_iter {
            let rss_channel = self.rss_feed_content_map.entry(
                String::from(feed_url)).or_insert(
                extract_rss_feed_channel_from_url(String::from(feed_url))?
            );
            loaded_rss_feed.push(rss_channel.clone());

        }

        self.rss_feeds = RssFeedList {
            items: loaded_rss_feed,
            state: ListState::default()
        };

        self.selected_rss_feed_tbl_state = TableState::default();
        self.current_view_state = AppState::FeedView;

        Ok(())

    }

    /// Run the application's main loop.
    pub fn run(mut self, mut terminal: DefaultTerminal) -> Result<()> {
        self.running = true;
        while self.running {

            terminal.draw(|frame| self.render(frame))?;
            self.handle_crossterm_events()?;

        }
        Ok(())
    }

    /// Renders the user interface.
    ///
    /// This is where you add new widgets. See the following resources for more information:
    ///
    /// - <https://docs.rs/ratatui/latest/ratatui/widgets/index.html>
    /// - <https://github.com/ratatui/ratatui/tree/main/ratatui-widgets/examples>
    fn render(&mut self, frame: &mut Frame) {

        let layout = layout::Layout::default()
            .direction(Direction::Horizontal)
            .constraints(vec![
                Constraint::Percentage(30),
                Constraint::Percentage(70)
            ])
            .split(frame.area());
            
        
        // The side tab for all of the RSS feeds: 
        let feed_iter = self.rss_feeds.items.iter();
        let mut list_items = Vec::<ListItem>::new();
        for feed in feed_iter {
            list_items.push(
                ListItem::new(
            Line::from(Span::raw(&feed.title))
                )
            );
        }

        // Populate with the list of RSS feeds
        let rss_feed_block =  List::new(list_items)
            .block(
                Block::new()
                    .borders(Borders::ALL)
                    .border_type(BorderType::Plain)
                    .title(format!("RSS Feeds"))
            )
            .highlight_symbol(">>");

        frame.render_stateful_widget(rss_feed_block, layout[0], & mut self.rss_feeds.state);

        let main_content_block = Block::new()
            .borders(Borders::ALL)
            .border_type(BorderType::Plain)
            .title(format!("Content"));
        
        match self.current_view_state {

            AppState::FeedView => {


                if let Some(selected_item) = self.rss_feeds.state.selected() {

                    let rss_channel = &self.rss_feeds.items[selected_item];
                    let rss_items_iter = rss_channel.items().iter();
                    
                    let mut item_rows= Vec::<Row>::new();
                    for item in rss_items_iter {

                        item_rows.push(Row::new(vec![
                            Cell::new(item.title.clone().unwrap_or_else(|| "-".to_string())),
                            Cell::new(item.pub_date.clone().unwrap_or_else(|| "-".to_string())),
                            Cell::new(item.description.clone().unwrap_or_else(|| "-".to_string()))
                        ]));
                    }

                    let width = [
                        Constraint::Percentage(30),
                        Constraint::Percentage(5),
                        Constraint::Percentage(65),
                    ];

                    let table = Table::new(item_rows, width)
                        .block(main_content_block)
                        .highlight_symbol(">");

                    frame.render_widget(table, layout[1]);

                } else {
                    frame.render_widget(main_content_block, layout[1]);
                }
            },

            AppState::ArticleView => {

                if let Some(selected_item) = self.rss_feeds.state.selected() {

                    let rss_channel = &self.rss_feeds.items[selected_item];
                    let rss_items_iter = rss_channel.items().iter();
                    
                    let mut item_rows= Vec::<Row>::new();
                    for item in rss_items_iter {

                        item_rows.push(Row::new(vec![
                            Cell::new(item.title.clone().unwrap_or_else(|| "-".to_string())),
                            Cell::new(item.pub_date.clone().unwrap_or_else(|| "-".to_string())),
                            Cell::new(item.description.clone().unwrap_or_else(|| "-".to_string()))
                        ]));
                    }

                    let width = [
                        Constraint::Percentage(30),
                        Constraint::Percentage(5),
                        Constraint::Percentage(65),
                    ];

                    let table = Table::new(item_rows, width)
                        .block(main_content_block)
                        .highlight_symbol(">");

                    frame.render_stateful_widget(table, layout[1], &mut self.selected_rss_feed_tbl_state);
                }
            },


            AppState::ReadingView => {

                // The direct channel:
                let selected_article = self.rss_feeds.state.selected().unwrap();
                let rss_channel = &self.rss_feeds.items[selected_article];

                let item_idx = self.selected_rss_feed_tbl_state.selected().unwrap();
                let rss_item= rss_channel.items().to_vec()[item_idx].clone();

                if let Some(article_url) = rss_item.link() {

                    let paragraph_txt = self.existing_articles_map.entry(
                        String::from(article_url)
                    ).or_insert(

                extract_full_content_from_url(String::from(article_url))
                            .unwrap_or_else(|e| e.to_string())

                    );
                    
                    self.vertical_scroll_state = self.vertical_scroll_state.content_length(
                        paragraph_txt.clone().len()
                    );
                    let rss_content = Paragraph::new(paragraph_txt.clone())
                        .block(main_content_block)
                        .alignment(layout::Alignment::Left)
                        .wrap(Wrap { trim: true})
                        .scroll((self.vertical_scroll as u16, 0));

                    frame.render_widget(rss_content, layout[1]);
                    frame.render_stateful_widget(
                        Scrollbar::new(ratatui::widgets::ScrollbarOrientation::VerticalRight),
                        layout[1], 
                        &mut self.vertical_scroll_state);
                }
            },

        }
    }

    /// Reads the crossterm events and updates the state of [`App`].
    ///
    /// If your application needs to perform work in between handling events, you can use the
    /// [`event::poll`] function to check if there are any events available with a timeout.
    fn handle_crossterm_events(&mut self) -> Result<()> {

        match event::read()? {

            // it's important to check KeyEventKind::Press to avoid handling key release events
            Event::Key(key) if key.kind == KeyEventKind::Press => self.on_key_event(key),
            Event::Mouse(_) => {}
            Event::Resize(_, _) => {}
            _ => {}
        }
        Ok(())
    }

    /// Handles the key events and updates the state of [`App`].
    fn on_key_event(&mut self, key: KeyEvent) {

        match (key.modifiers, key.code) {

            (_, KeyCode::Esc | KeyCode::Char('q'))
            | (KeyModifiers::CONTROL, KeyCode::Char('c') | KeyCode::Char('C')) => self.quit(),

            (_, KeyCode::Char('j')) => {

                match self.current_view_state {

                    AppState::FeedView => {
                        self.rss_feeds.state.select_next()
                    },

                    AppState::ArticleView => {
                        self.selected_rss_feed_tbl_state.select_next();
                    },

                    AppState::ReadingView => {
                        self.vertical_scroll = self.vertical_scroll.saturating_add(1);
                        self.vertical_scroll_state = self.vertical_scroll_state.position(self.vertical_scroll);
                    },

                }
            }

            (_, KeyCode::Char('k')) => {

                match self.current_view_state {

                    AppState::FeedView => {
                        self.rss_feeds.state.select_previous();
                    },

                    AppState::ArticleView => {
                        self.selected_rss_feed_tbl_state.select_previous();
                    },

                    AppState::ReadingView => {
                        self.vertical_scroll = self.vertical_scroll.saturating_sub(1);
                        self.vertical_scroll_state = self.vertical_scroll_state.position(self.vertical_scroll)
                    },

                }
            }


            (_, KeyCode::Enter | KeyCode::Char('l')) => {

                match self.current_view_state {

                    AppState::FeedView => {
                        self.current_view_state = AppState::ArticleView;
                        self.selected_rss_feed_tbl_state.select_first();
                    },

                    AppState::ArticleView => {
                        self.current_view_state = AppState::ReadingView;
                        self.vertical_scroll = 0;
                        self.vertical_scroll_state = self.vertical_scroll_state.position(self.vertical_scroll)
                    },

                    AppState::ReadingView => todo!(),

                }
            }

            (_, KeyCode::Char('b') | KeyCode::Char('h')) => {

                match self.current_view_state {

                    AppState::FeedView => {},

                    AppState::ArticleView => {
                        self.current_view_state = AppState::FeedView;
                    },

                    AppState::ReadingView => {
                        self.current_view_state = AppState::ArticleView;
                    },

                }
            }

            _ => {}
        }
    }

    /// Set running to false to quit the application.
    fn quit(&mut self) {
        self.running = false;
    }
}
