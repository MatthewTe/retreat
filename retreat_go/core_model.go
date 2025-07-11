package main

import (
	"github.com/MatthewTe/retreat/database"

	key "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RetreatModel struct {
	DBPath string

	// Article management vars:
	ArticleList   list.Model
	SelectedTitle string
}

func (m RetreatModel) Init() tea.Cmd {
	// On init use the DBpath to trigger inital load
	return nil
}

func (m RetreatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case database.DatabaseArticlesMsg:
		{
			m.ArticleList = list.New(msg, list.NewDefaultDelegate(), 0, 0)
			m.ArticleList.Title = "Retreat RSS Feeds"
			return m, nil
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			{
				return m, tea.Quit
			}

		case key.Matches(msg, DefaultKeyMap.Up):
		case key.Matches(msg, DefaultKeyMap.Down):
		}

	case tea.WindowSizeMsg:
		{

			m.ArticleList.SetSize(msg.Width, msg.Height)
		}

	}

	// Pass the message through to the list component:
	var cmd tea.Cmd
	m.ArticleList, cmd = m.ArticleList.Update(msg)
	return m, cmd

}

func (m RetreatModel) View() string {
	return m.ArticleList.View()
}
