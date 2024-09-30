package agent

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
)

const (
	requestSessionKeyItem = iota
	writeMessageItem
)

var _ tea.Model = (*Agent)(nil)

var (
	activeStyle   = lip.NewStyle().Foreground(lip.Color("255"))
	inactiveStyle = lip.NewStyle().Foreground(lip.Color("240"))
)

type Agent struct {
	cfg *config
	*tui
}

type tui struct {
	items  []string
	active map[int]struct{}
	cursor int
}

func NewAgent() *Agent {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	return &Agent{
		cfg: cfg,
		tui: initialTUI(),
	}
}

func initialTUI() *tui {
	return &tui{
		items: []string{
			"Request session key",
			"Write a message",
		},
		active: map[int]struct{}{
			requestSessionKeyItem: {},
		},
		cursor: requestSessionKeyItem,
	}
}

func (a Agent) Init() tea.Cmd {
	return nil
}

func (a Agent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "up":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down":
			_, ok := a.tui.active[a.cursor+1]
			if a.cursor < len(a.items)-1 && ok {
				a.cursor++
			}
		case "enter":
			return a, selectItemCmd(a.cursor, a.active)
		}
	case SessionKeyMsg:

	case ErrorMsg:

	}

	return a, nil
}

func (a Agent) View() string {
	var s string

	for i, item := range a.items {
		cursor := " "
		if a.cursor == i {
			cursor = ">"
		}

		style := inactiveStyle
		if _, ok := a.active[i]; ok {
			style = activeStyle
		}

		s += fmt.Sprintf("%s %s\n", cursor, style.Render(item))
	}

	s += "\nPress q to quit\n"

	return s
}

func (a Agent) Run() {
	prog := tea.NewProgram(a, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		// TODO: Add logging
	}
}

// Cmd

func selectItemCmd(cursor int, active map[int]struct{}) tea.Cmd {
	return func() tea.Msg {
		switch cursor {
		case requestSessionKeyItem:
			// TODO: requestSessionKey()
			active[writeMessageItem] = struct{}{}
			return SessionKeyMsg([]byte("dummy"))
		case 1:
			// TODO: writeMessage()
			return ErrorMsg(nil)
		}
		return ErrorMsg(nil)
	}
}

// Msg

type SessionKeyMsg []byte

type ErrorMsg error
