package agent

import tea "github.com/charmbracelet/bubbletea"

type Agent struct {
	model *cliModel
}

func NewAgent() *Agent {
	return &Agent{
		model: NewModel(),
	}
}

func (a Agent) Run() {
	prog := tea.NewProgram(a.model)
	if _, err := prog.Run(); err != nil {
		// TODO: Add logging
	}
}
