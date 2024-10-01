package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
	"github.com/sudeeya/key-exchange/internal/pkg/api"
	"github.com/sudeeya/key-exchange/internal/pkg/crypto"
	"github.com/sudeeya/key-exchange/internal/pkg/pem"
	"github.com/sudeeya/key-exchange/internal/pkg/rng"
)

const (
	menuMode = iota
	requestMode
	messageMode
)

const (
	requestSessionKeyItem = iota
	writeMessageItem
)

const httpPrefix = "http://"

var _ tea.Model = (*Agent)(nil)

var (
	activeStyle   = lip.NewStyle().Foreground(lip.Color("255"))
	inactiveStyle = lip.NewStyle().Foreground(lip.Color("240"))
	staticStyle   = lip.NewStyle().Foreground(lip.Color("133"))
)

type Agent struct {
	cfg    *config
	tui    *tui
	keys   *keys
	client *resty.Client
	rng    *rng.RNG
}

type tui struct {
	mode     int
	items    []string
	active   map[int]struct{}
	cursor   int
	input    textinput.Model
	sessions []string
}

type keys struct {
	privateKey []byte
	//publicKey   []byte
	trentKey []byte
	//agentKeys   map[string][]byte
	//sessionKeys map[string][]byte
}

func NewAgent() *Agent {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	keys, err := initialKeys(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &Agent{
		cfg:    cfg,
		tui:    initialTUI(),
		keys:   keys,
		client: resty.New(),
		rng:    rng.NewRNG(),
	}
}

func initialTUI() *tui {
	return &tui{
		mode: menuMode,
		items: []string{
			"Request session key",
			"Write a message",
		},
		active: map[int]struct{}{
			requestSessionKeyItem: {},
		},
		cursor:   requestSessionKeyItem,
		input:    textinput.New(),
		sessions: make([]string, 0),
	}
}

func initialKeys(cfg *config) (*keys, error) {
	privateKey, err := pem.ExtractRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	trentKey, err := pem.ExtractRSAPublicKey(cfg.TrentPublicKey)
	if err != nil {
		return nil, err
	}

	return &keys{
		privateKey: privateKey,
		trentKey:   trentKey,
	}, nil
}

func (a Agent) Init() tea.Cmd {
	return nil
}

func (a Agent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			switch a.tui.mode {
			case menuMode:
				return a, tea.Quit
			}
		case "up":
			switch a.tui.mode {
			case menuMode:
				if a.tui.cursor > 0 {
					a.tui.cursor--
				}
			}
		case "down":
			switch a.tui.mode {
			case menuMode:
				_, ok := a.tui.active[a.tui.cursor+1]
				if a.tui.cursor < len(a.tui.items)-1 && ok {
					a.tui.cursor++
				}
			}
		case "enter":
			switch a.tui.mode {
			case menuMode:
				a.tui.input.Focus()
				return a, selectItemCmd(a.tui)
			case requestMode:
				agentID := a.tui.input.Value()
				a.tui.input.Reset()

				if !slices.Contains(a.cfg.AgentIDs, agentID) {
					a.tui.input.Placeholder = "Agent with such ID does not exist. Try again"
					return a, nil
				}

				a.tui.input.Blur()
				a.tui.mode = menuMode

				return a, requestSessionKey(a.cfg, a.keys, a.client, a.rng, agentID)
			case messageMode:

			}
		}
	case ModeMsg:
		a.tui.mode = int(msg)
	case SessionEstablishedMsg:
		a.tui.active[writeMessageItem] = struct{}{}
		a.tui.sessions = append(a.tui.sessions, string(msg))
	case ErrorMsg:
		log.Println(msg)
	}

	switch a.tui.mode {
	case requestMode, messageMode:
		var cmd tea.Cmd
		a.tui.input, cmd = a.tui.input.Update(msg)
		return a, cmd
	case menuMode:
		return a, nil
	}

	return a, nil
}

func (a Agent) View() string {
	var s strings.Builder

	switch a.tui.mode {
	case menuMode:
		for i, item := range a.tui.items {
			cursor := " "
			if a.tui.cursor == i {
				cursor = ">"
			}

			style := inactiveStyle
			if _, ok := a.tui.active[i]; ok {
				style = activeStyle
			}

			s.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(item)))

		}

		for _, session := range a.tui.sessions {
			s.WriteString(staticStyle.Render(fmt.Sprintf("\nSession with %s established\n", session)))
		}

		s.WriteString(staticStyle.Render("\nPress q to quit\n"))
	case requestMode, messageMode:
		s.WriteString(a.tui.input.View())
	}

	return s.String()
}

func (a Agent) Run() {
	prog := tea.NewProgram(a, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		// TODO: Add logging
	}
}

// Cmd

func selectItemCmd(tui *tui) tea.Cmd {
	return func() tea.Msg {
		switch tui.cursor {
		case requestSessionKeyItem:
			tui.input.Placeholder = "Enter agent ID"
			return ModeMsg(requestMode)
		case writeMessageItem:
			tui.input.Placeholder = "Enter your message"
			return ModeMsg(messageMode)
		}

		return ModeMsg(menuMode)
	}
}

func requestSessionKey(cfg *config, keys *keys, client *resty.Client, rng *rng.RNG, acceptor string) tea.Cmd {
	return func() tea.Msg {
		// Step 1
		req1 := api.Request{
			Initiator: cfg.ID,
			Acceptor:  acceptor,
		}
		var resp1 api.Response
		_, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req1).
			SetResult(&resp1).
			Post(httpPrefix + cfg.TrentAddr + api.Step1Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}

		infoJSON, err := json.Marshal(resp1.Certificate.Information)
		if err != nil {
			return ErrorMsg(err)
		}
		ok := crypto.VerifyRSA(infoJSON, resp1.Certificate.Signature, keys.trentKey)
		if !ok {
			return ErrorMsg(fmt.Errorf("signature verification failed"))
		}

		acceptorKey := resp1.Certificate.Information.AcceptorKey

		// Step 3
		initiatorNonce, err := rng.GenerateNonce()
		if err != nil {
			return ErrorMsg(err)
		}
		info3 := api.Info{
			Initiator:      cfg.ID,
			InitiatorNonce: initiatorNonce,
		}
		infoJSON3, err := json.Marshal(info3)
		if err != nil {
			return ErrorMsg(err)
		}
		ciphertext := crypto.EncryptRSA(infoJSON3, acceptorKey)

		acceptorAddr := getAgentAddr(acceptor, cfg)
		_, err = client.R().
			SetHeader("Content-Type", "text/plain").
			SetBody(ciphertext).
			Post(httpPrefix + acceptorAddr + api.Step3Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}

		return SessionEstablishedMsg(acceptor)
	}
}

// Msg

type ModeMsg int

type SessionEstablishedMsg string

type ErrorMsg error
