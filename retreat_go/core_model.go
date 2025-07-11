package main

import (
	"log"

	"github.com/MatthewTe/retreat/database"

	key "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type AppState int

const (
	ArticleListState AppState = iota
	ArticleMarkdownState
)

type RetreatModel struct {
	DBPath string
	State  AppState

	// Article management vars:
	ArticleList     list.Model
	ArticleViewPort viewport.Model
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

	// Can I do a switch statement based on the State? What is the best way to do this?
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			{
				return m, tea.Quit
			}

		case key.Matches(msg, DefaultKeyMap.Select):

			switch m.State {
			case ArticleListState:
				{
					if !(m.ArticleList.FilterState() == list.Filtering) {
						article, ok := m.ArticleList.SelectedItem().(database.ArticleItem)
						if ok {
							articleMarkdown, err := database.GetArticleMarkdownContent(m.DBPath, article.ArticleTitle)
							if err != nil {
								log.Fatal(err)
							}
							m.ArticleViewPort.SetContent(articleMarkdown)
							m.State = ArticleMarkdownState
						}
					}
				}
			}

		case key.Matches(msg, DefaultKeyMap.Back):
			{
				switch m.State {
				case ArticleMarkdownState:
					{
						m.State = ArticleListState
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		{

			m.ArticleList.SetSize(msg.Width, msg.Height)
			m.ArticleViewPort = viewport.New(msg.Width, msg.Height)
			//m.ArticleViewPort.YPosition = 0
		}

	}

	// Pass the message through to the list component:
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.ArticleViewPort, cmd = m.ArticleViewPort.Update(msg)
	cmds = append(cmds, cmd)

	m.ArticleList, cmd = m.ArticleList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)

}

func (m RetreatModel) View() string {

	switch m.State {
	case ArticleListState:
		return m.ArticleList.View()
	case ArticleMarkdownState:
		return m.ArticleViewPort.View()
	}

	return m.ArticleList.View()

}
