package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lip "github.com/charmbracelet/lipgloss"
	"github.com/go-chi/chi/v5"
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
	errorStyle    = lip.NewStyle().Foreground(lip.Color("160"))
)

type Agent struct {
	cfg    *config
	tui    *tui
	keys   *keys
	client *resty.Client
	mux    *chi.Mux
	rng    *rng.RNG
	msgCh  chan tea.Msg
}

type tui struct {
	mode     int
	items    []string
	active   map[int]struct{}
	cursor   int
	input    textinput.Model
	sessions []string
	err      string
}

type keys struct {
	privateKey []byte
	trentKey   []byte
	agentKey   []byte
	sessionKey []byte
	nonce      []byte
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

	mux := chi.NewRouter()

	return &Agent{
		cfg:    cfg,
		tui:    initialTUI(),
		keys:   keys,
		client: resty.New(),
		mux:    mux,
		rng:    rng.NewRNG(),
		msgCh:  make(chan tea.Msg),
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
		err:      "",
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
		agentKey:   make([]byte, 1),
		sessionKey: make([]byte, 1),
	}, nil
}

func (a Agent) Init() tea.Cmd {
	return checkMsgChanCmd(&a)
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
		case "esc":
			switch a.tui.mode {
			case requestMode, messageMode:
				a.tui.mode = menuMode
				a.tui.input.Reset()
				a.tui.input.Blur()
				return a, nil
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

				if agentID != a.cfg.AgentID {
					a.tui.input.Placeholder = "Agent with such ID does not exist. Try again"
					return a, nil
				}

				a.tui.input.Blur()
				a.tui.mode = menuMode

				return a, requestSessionKeyCmd(a.cfg, a.keys, a.client, a.rng, agentID)
			case messageMode:

			}
		}
	case ModeMsg:
		a.tui.mode = int(msg)
	case SessionEstablishedMsg:
		a.tui.active[writeMessageItem] = struct{}{}
		a.tui.sessions = append(a.tui.sessions, string(msg))
		return a, nil
	case ErrorMsg:
		a.tui.err = error(msg).Error()
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
	s.WriteString("\n")

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

			s.WriteString(fmt.Sprintf(" %s %s\n", activeStyle.Render(cursor), style.Render(item)))

		}

		if a.tui.err != "" {
			s.WriteString(errorStyle.Render(fmt.Sprintf("\n %s\n", a.tui.err)))
		}

		for _, session := range a.tui.sessions {
			s.WriteString(inactiveStyle.Render(fmt.Sprintf("\n Session with %s established\n", session)))
		}

		s.WriteString(inactiveStyle.Render("\n Press q to quit\n"))
	case requestMode, messageMode:
		s.WriteString(" " + a.tui.input.View() + "\n")

		s.WriteString(inactiveStyle.Render("\n Press esc to return to the menu\n"))
	}

	return s.String()
}

func (a Agent) Run() {
	a.addRoutes()

	go func() {
		if err := http.ListenAndServe(a.cfg.Addr, a.mux); err != nil {
			log.Fatal(err)
		}
	}()

	prog := tea.NewProgram(a, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *Agent) addRoutes() {
	a.mux.Post(api.Step4Endpoint, step4Handler(a))
	a.mux.Post(api.Step7Endpoint, step7Handler(a))
}

// Cmd

func checkMsgChanCmd(a *Agent) tea.Cmd {
	go func() {
		for msg := range a.msgCh {
			a.Update(msg)
		}
	}()

	return nil
}

func selectItemCmd(tui *tui) tea.Cmd {
	return func() tea.Msg {
		tui.err = ""

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

func requestSessionKeyCmd(cfg *config, keys *keys, client *resty.Client, rng *rng.RNG, acceptor string) tea.Cmd {
	return func() tea.Msg {
		// Step 1
		req1 := api.Request{
			Initiator: cfg.ID,
			Acceptor:  acceptor,
		}
		var resp2 api.Response
		rawResp2, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req1).
			SetResult(&resp2).
			Post(httpPrefix + cfg.TrentAddr + api.Step2Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}
		if rawResp2.StatusCode() != http.StatusOK {
			return ErrorMsg(fmt.Errorf("step 2 status code is %d", rawResp2.StatusCode()))
		}

		infoJSON, err := json.Marshal(resp2.Certificate.Information)
		if err != nil {
			return ErrorMsg(err)
		}
		ok := crypto.VerifyRSA(infoJSON, resp2.Certificate.Signature, keys.trentKey)
		if !ok {
			return ErrorMsg(fmt.Errorf("signature verification failed"))
		}

		keys.agentKey = resp2.Certificate.Information.AcceptorKey

		// Step 3
		initiatorNonce, err := rng.GenerateNonce()
		if err != nil {
			return ErrorMsg(err)
		}
		keys.nonce = initiatorNonce
		info3 := api.Info{
			Initiator:      cfg.ID,
			InitiatorNonce: initiatorNonce,
		}
		infoJSON3, err := json.Marshal(info3)
		if err != nil {
			return ErrorMsg(err)
		}
		ciphertext := crypto.EncryptRSA(infoJSON3, keys.agentKey)

		req3 := api.Request{
			Ciphertext: ciphertext,
		}
		acceptorAddr := cfg.AgentAddr
		var resp4 api.Response
		rawResp3, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req3).
			SetResult(&resp4).
			Post(httpPrefix + acceptorAddr + api.Step4Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}
		if rawResp3.StatusCode() != http.StatusOK {
			return ErrorMsg(fmt.Errorf("step 4 status code is %d", rawResp3.StatusCode()))
		}

		respJSON := crypto.DecryptRSA(resp4.Ciphertext, keys.privateKey)

		var resp api.Response
		if err := json.Unmarshal(respJSON, &resp); err != nil {
			return ErrorMsg(err)
		}

		keys.sessionKey = resp.Certificate.Information.SessionKey

		// Step 7
		iv, err := rng.GenerateIV()
		if err != nil {
			return ErrorMsg(err)
		}
		ciphertext7 := crypto.EncryptAES(resp.AcceptorNonce, keys.sessionKey, iv)
		msg := api.Message{
			IV:         iv,
			Ciphertext: ciphertext7,
		}
		rawResp7, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(msg).
			Post(httpPrefix + acceptorAddr + api.Step7Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}
		if rawResp7.StatusCode() != http.StatusOK {
			return ErrorMsg(fmt.Errorf("step 7 status code is %d", rawResp7.StatusCode()))
		}

		return SessionEstablishedMsg(acceptor)
	}
}

// Msg

type ModeMsg int

type SessionEstablishedMsg string

type ErrorMsg error
