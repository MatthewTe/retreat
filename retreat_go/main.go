package main

import (
	"flag"
	"log"

	"github.com/MatthewTe/retreat/database"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	dbPath := flag.String("database", "./retreat.db", "The Database that is used to render the articles")
	tryToSync := flag.Bool("sync", false, "If it will check the s3 bucket for a new db to sync")

	flag.Parse()

	if *tryToSync {
		err := database.SyncDatabsaeFromS3(*dbPath)
		if err != nil {
			log.Fatalln(err)
		}

	}

	var articles []list.Item = database.LoadArticlesFromBlob(*dbPath)
	var feeds []list.Item = database.LoadFeedsFromDB(*dbPath)
	m := RetreatModel{
		DBPath:        *dbPath,
		CommandPallet: textinput.New(),

		ArticleList: list.New(articles, list.NewDefaultDelegate(), 0, 0),
		FeedList:    list.New(feeds, list.NewDefaultDelegate(), 0, 0),
		State:       FeedListState,
	}

	// Initalize the Bubbletea TUI:
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalln(err)
	}

}
