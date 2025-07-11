package main

import (
	"log"

	"github.com/MatthewTe/retreat/database"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// For go server these things need to get done:
/*
- I provide it with aws config ( or load aws config onto machine)
- It checks the last uploaded date of the db in blob. If its newer than the local copy it pulls it down into preset dir.
- It opens the TUI w/ the basic read functions for the two tables and renders a formatted table view of all feed articles.
- When an article is clicked on it renders the markdown content as a scrollable page via https://github.com/charmbracelet/glamour?tab=readme-ov-file
*/

var (
	defaultDBPath = "./retreat.db"
)

func main() {

	var articles []list.Item = database.LoadFileFromBlob(defaultDBPath)
	m := RetreatModel{DBPath: defaultDBPath, ArticleList: list.New(articles, list.NewDefaultDelegate(), 0, 0)}

	// Initalize the Bubbletea TUI:
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalln(err)
	}

}
