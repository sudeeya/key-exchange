package agent

import (
	tea "github.com/charmbracelet/bubbletea"
)

var _ tea.Model = (*cliModel)(nil)

type cliModel struct {
}

func newModel() *cliModel {
	return &cliModel{}
}

func (m cliModel) Init() tea.Cmd {
	return nil
}

func (m *cliModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m cliModel) View() string {
	return "\nPress q to quit\n"
}
