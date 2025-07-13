package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/MatthewTe/retreat/database"
	key "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type AppState int

const (
	ArticleListState AppState = iota
	FeedListState
	ArticleMarkdownState
	userCmdInputState
)

type RetreatModel struct {
	Ready bool

	DBPath string
	State  AppState

	// Article management vars:
	ArticleList     list.Model
	FeedList        list.Model
	ArticleViewPort viewport.Model

	// Footer application bar:
	CommandPallet textinput.Model
}

func (m RetreatModel) Init() tea.Cmd {
	// On init use the DBpath to trigger inital load
	return nil
}

func (m RetreatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, DefaultKeyMap.CmdPallet):
			{
				m.State = userCmdInputState
				m.CommandPallet.Focus()
			}

		case key.Matches(msg, DefaultKeyMap.Quit):
			{
				return m, tea.Quit
			}

		case key.Matches(msg, DefaultKeyMap.Select):

			switch m.State {

			case FeedListState:
				{
					if !(m.FeedList.FilterState() == list.Filtering) {
						feed, ok := m.FeedList.SelectedItem().(database.FeedItem)
						if ok {
							articlesFromFeed := database.GetArticlesFromFeed(m.DBPath, feed.FeedTitle)
							m.ArticleList = list.New(articlesFromFeed, list.NewDefaultDelegate(), m.ArticleList.Width(), m.ArticleList.Height())
						}

						m.State = ArticleListState
					}
				}

			case ArticleListState:
				{
					if !(m.ArticleList.FilterState() == list.Filtering) {
						article, ok := m.ArticleList.SelectedItem().(database.ArticleItem)
						if ok {
							articleMarkdown, err := database.GetArticleMarkdownContent(m.DBPath, article.ArticleTitle)
							if err != nil {
								log.Fatal(err)
							}
							content, err := glamour.Render(articleMarkdown, "dark")
							if err != nil {
								log.Fatal(err)
							}

							m.ArticleViewPort.SetContent(content)
							m.State = ArticleMarkdownState
						}
					}
				}

			case userCmdInputState:
				{
					// Implement Commands here that I can pipe in to other stuff
					m.CommandPallet.SetValue("")
					m.State = FeedListState
				}

			}

		case key.Matches(msg, DefaultKeyMap.Back):
			{
				switch m.State {
				case ArticleListState:
					{
						m.State = FeedListState
					}

				case ArticleMarkdownState:
					{
						m.State = ArticleListState
					}
				}
			}
		case key.Matches(msg, DefaultKeyMap.Left):
			{
				switch m.State {
				case FeedListState:
					{
						m.State = ArticleListState
					}
				}
			}

		case key.Matches(msg, DefaultKeyMap.Right):
			{
				switch m.State {
				case ArticleListState:
					{
						m.State = FeedListState
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		{

			headerHeight := lipgloss.Height(m.headerView())
			footerHeight := lipgloss.Height(m.footerView())
			verticalMarginHeight := headerHeight + footerHeight

			if !m.Ready {
				m.ArticleList.SetSize(msg.Width/2, msg.Height/2)
				m.FeedList.SetSize(msg.Width/2, msg.Height/2)
				m.ArticleViewPort.Height = msg.Height - verticalMarginHeight
				m.ArticleViewPort = viewport.New(msg.Width, msg.Height)
				m.ArticleViewPort.YPosition = headerHeight
				m.Ready = true
			} else {
				m.ArticleList.SetSize(msg.Width/2, msg.Height/2)
				m.FeedList.SetSize(msg.Width/2, msg.Height/2)
				m.ArticleViewPort = viewport.New(msg.Width, msg.Height)
				m.ArticleViewPort = viewport.New(msg.Width, msg.Height)
				m.ArticleViewPort.Width = msg.Width
				m.ArticleViewPort.Height = msg.Height - verticalMarginHeight
			}
		}
	}

	// Pass the message through to the list component:
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Only passing in commands to the component focused on:
	switch m.State {
	case FeedListState:
		{
			m.FeedList, cmd = m.FeedList.Update(msg)
			cmds = append(cmds, cmd)
		}
	case ArticleListState:
		{
			m.ArticleList, cmd = m.ArticleList.Update(msg)
			cmds = append(cmds, cmd)
		}
	case ArticleMarkdownState:
		{
			m.ArticleViewPort, cmd = m.ArticleViewPort.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	m.CommandPallet, cmd = m.CommandPallet.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)

}

func (m RetreatModel) View() string {

	totalViewString := ""

	feedListView := m.FeedList.View()
	articleListView := m.ArticleList.View()

	cmdPalletView := cmdPalletUnFocusStyle.Render(m.CommandPallet.View())

	switch m.State {
	case FeedListState:

		totalViewString += lipgloss.JoinHorizontal(
			lipgloss.Left,
			focusedListStyle.Render(feedListView),
			unfocusedListStyle.Render(articleListView),
		)

	case ArticleListState:

		totalViewString += lipgloss.JoinHorizontal(
			lipgloss.Left,
			unfocusedListStyle.Render(feedListView),
			focusedListStyle.Render(articleListView),
		)

	case ArticleMarkdownState:
		totalViewString = fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.ArticleViewPort.View(), m.footerView())

	case userCmdInputState:
		cmdPalletView = cmdPalletFocusStyle.Render(cmdPalletView)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		totalViewString,
		cmdPalletView,
	)

}
func (m RetreatModel) headerView() string {
	title := titleStyle.Render("Example Title")
	line := strings.Repeat("─", max(0, m.ArticleViewPort.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m RetreatModel) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.ArticleViewPort.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.ArticleViewPort.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
