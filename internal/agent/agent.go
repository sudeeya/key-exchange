package agent

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type Agent struct {
	cfg   *config
	model *cliModel
}

func NewAgent() *Agent {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	return &Agent{
		cfg:   cfg,
		model: NewModel(),
	}
}

func (a Agent) Run() {
	prog := tea.NewProgram(a.model)
	if _, err := prog.Run(); err != nil {
		// TODO: Add logging
	}
}
