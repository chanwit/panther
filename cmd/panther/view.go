package main

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type simpleMarkdownView struct {
	viewport viewport.Model
}

func newExample(content string) (*simpleMarkdownView, error) {
	width, height, err := term.GetSize(0)
	if err != nil {
		return nil, err
	}
	width = width - 1
	height = height - 2

	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-5),
	)
	if err != nil {
		return nil, err
	}

	str, err := renderer.Render(content)
	if err != nil {
		return nil, err
	}

	vp.SetContent(str)

	return &simpleMarkdownView{
		viewport: vp,
	}, nil
}

func (e simpleMarkdownView) Init() tea.Cmd {
	return nil
}

func (e simpleMarkdownView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return e, tea.Quit
		default:
			var cmd tea.Cmd
			e.viewport, cmd = e.viewport.Update(msg)
			return e, cmd
		}
	default:
		return e, nil
	}
}

func (e simpleMarkdownView) View() string {
	return e.viewport.View() + e.helpView()
}

func (e simpleMarkdownView) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}
