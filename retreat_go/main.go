package main

// For go server these things need to get done:
/*
- I provide it with aws config ( or load aws config onto machine)
- It checks the last uploaded date of the db in blob. If its newer than the local copy it pulls it down into preset dir.
- It opens the TUI w/ the basic read functions for the two tables and renders a formatted table view of all feed articles.
- When an article is clicked on it renders the markdown content as a scrollable page via https://github.com/charmbracelet/glamour?tab=readme-ov-file
*/

func main() {

	RunDevRSSServer()

}
