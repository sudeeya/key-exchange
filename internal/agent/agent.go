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
	"go.uber.org/zap"
)

const (
	menuMode = iota
	requestMode
	mailMode
	messageMode
)

const (
	requestSessionKeyItem = iota
	mailboxItem
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
	logger *zap.Logger
	tui    *tui
	keys   *keys
	client *resty.Client
	mux    *chi.Mux
	rng    *rng.RNG
}

type tui struct {
	mode     int
	items    []string
	active   map[int]struct{}
	cursor   int
	input    textinput.Model
	session  string
	messages []string
	unread   bool
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

	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.OutputPaths = []string{
		cfg.LogFile,
	}
	logger, err := loggerCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Initializing keys")
	keys, err := initialKeys(cfg)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Initializing router")
	mux := chi.NewRouter()

	return &Agent{
		cfg:    cfg,
		logger: logger,
		tui:    initialTUI(),
		keys:   keys,
		client: resty.New(),
		mux:    mux,
		rng:    rng.NewRNG(),
	}
}

func initialTUI() *tui {
	return &tui{
		mode: menuMode,
		items: []string{
			"Request session key",
			"Mailbox",
			"Write a message",
		},
		active: map[int]struct{}{
			requestSessionKeyItem: {},
			mailboxItem:           {},
		},
		cursor:   requestSessionKeyItem,
		input:    textinput.New(),
		session:  "",
		messages: make([]string, 0),
		unread:   false,
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
		agentKey:   make([]byte, 0),
		sessionKey: make([]byte, 0),
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
		case "esc":
			switch a.tui.mode {
			case requestMode, messageMode:
				a.tui.mode = menuMode
				a.tui.input.Reset()
				a.tui.input.Blur()
				return a, nil
			case mailMode:
				a.tui.unread = false
				a.tui.mode = menuMode
				return a, nil
			}
		case "up":
			switch a.tui.mode {
			case menuMode:
				if a.tui.cursor > 0 {
					a.tui.cursor--
				}
				return a, nil
			}
		case "down":
			switch a.tui.mode {
			case menuMode:
				_, ok := a.tui.active[a.tui.cursor+1]
				if a.tui.cursor < len(a.tui.items)-1 && ok {
					a.tui.cursor++
				}
				return a, nil
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

				return a, requestSessionKeyCmd(&a, agentID)
			case messageMode:
				msg := a.tui.input.Value()
				a.tui.input.Reset()
				a.tui.input.Blur()
				a.tui.mode = menuMode

				return a, sendMessageCmd(&a, msg)
			}
		}
	case ModeChangedMsg:
		a.tui.mode = int(msg)
	case ErrorMsg:
		if error(msg) != nil {
			a.tui.err = error(msg).Error()
		}
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

			switch {
			case i == mailboxItem && a.tui.unread:
				s.WriteString(fmt.Sprintf(" %s [!] %s\n", activeStyle.Render(cursor), style.Render(item)))
			default:
				s.WriteString(fmt.Sprintf(" %s     %s\n", activeStyle.Render(cursor), style.Render(item)))
			}

		}

		if a.tui.err != "" {
			s.WriteString(errorStyle.Render(fmt.Sprintf("\n %s\n", a.tui.err)))
		}

		if a.tui.session != "" {
			s.WriteString(inactiveStyle.Render(fmt.Sprintf("\n Session with %s established\n", a.tui.session)))
		}

		s.WriteString(inactiveStyle.Render("\n Press q to quit\n"))
	case requestMode, messageMode:
		s.WriteString(" " + a.tui.input.View() + "\n")

		s.WriteString(inactiveStyle.Render("\n Press esc to return to the menu\n"))
	case mailMode:
		if len(a.tui.messages) == 0 {
			s.WriteString(inactiveStyle.Render(" Mailbox is empty\n"))
		}

		for _, message := range a.tui.messages {
			point := "*"
			s.WriteString(fmt.Sprintf(" %s %s\n", activeStyle.Render(point), activeStyle.Render(message)))
		}

		s.WriteString(inactiveStyle.Render("\n Press esc to return to the menu\n"))
	}

	return s.String()
}

func (a Agent) Run() {
	a.logger.Info("Initializing endpoints")
	a.addRoutes()

	a.logger.Info("Agent is running")
	go func() {
		if err := http.ListenAndServe(a.cfg.Addr, a.mux); err != nil {
			a.logger.Fatal(err.Error())
		}
	}()

	prog := tea.NewProgram(a, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		a.logger.Fatal(err.Error())
	}
}

func (a *Agent) addRoutes() {
	a.mux.Post(api.Step4Endpoint, step4Handler(a))
	a.mux.Post(api.Step7Endpoint, step7Handler(a))
	a.mux.Post(api.MessageEndpoint, messageHandler(a))
}

// Cmd

func selectItemCmd(tui *tui) tea.Cmd {
	return func() tea.Msg {
		tui.err = ""

		switch tui.cursor {
		case requestSessionKeyItem:
			tui.input.Placeholder = "Enter agent ID"
			return ModeChangedMsg(requestMode)
		case mailboxItem:
			return ModeChangedMsg(mailMode)
		case writeMessageItem:
			tui.input.Placeholder = "Enter your message"
			return ModeChangedMsg(messageMode)
		}

		return ModeChangedMsg(menuMode)
	}
}

func requestSessionKeyCmd(a *Agent, acceptor string) tea.Cmd {
	return func() tea.Msg {
		// Step 1
		req1 := api.Request{
			Initiator: a.cfg.ID,
			Acceptor:  acceptor,
		}
		var resp2 api.Response
		rawResp2, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req1).
			SetResult(&resp2).
			Post(httpPrefix + a.cfg.TrentAddr + api.Step2Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}
		if rawResp2.StatusCode() != http.StatusOK {
			return ErrorMsg(fmt.Errorf("step 2 status code is %d", rawResp2.StatusCode()))
		}

		info2JSON, err := json.Marshal(resp2.Certificate.Information)
		if err != nil {
			return ErrorMsg(err)
		}
		ok := crypto.VerifyRSA(info2JSON, resp2.Certificate.Signature, a.keys.trentKey)
		if !ok {
			return ErrorMsg(fmt.Errorf("signature verification failed"))
		}

		a.keys.agentKey = resp2.Certificate.Information.AcceptorKey

		// Step 3
		initiatorNonce, err := a.rng.GenerateNonce()
		if err != nil {
			return ErrorMsg(err)
		}
		a.keys.nonce = initiatorNonce
		info3 := api.Info{
			Initiator:      a.cfg.ID,
			InitiatorNonce: initiatorNonce,
		}
		info3JSON, err := json.Marshal(info3)
		if err != nil {
			return ErrorMsg(err)
		}
		ciphertext3 := crypto.EncryptRSA(info3JSON, a.keys.agentKey)

		req3 := api.Request{
			Ciphertext: ciphertext3,
		}
		acceptorAddr := a.cfg.AgentAddr
		var resp4 api.Response
		rawResp4, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(req3).
			SetResult(&resp4).
			Post(httpPrefix + acceptorAddr + api.Step4Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}
		if rawResp4.StatusCode() != http.StatusOK {
			return ErrorMsg(fmt.Errorf("step 4 status code is %d", rawResp4.StatusCode()))
		}

		resp4JSON := crypto.DecryptRSA(resp4.Ciphertext, a.keys.privateKey)

		var resp api.Response
		if err := json.Unmarshal(resp4JSON, &resp); err != nil {
			return ErrorMsg(err)
		}

		a.keys.sessionKey = resp.Certificate.Information.SessionKey

		// Step 7
		iv, err := a.rng.GenerateIV()
		if err != nil {
			return ErrorMsg(err)
		}
		ciphertext7 := crypto.EncryptAES(resp.AcceptorNonce, a.keys.sessionKey, iv)
		msg := api.Message{
			IV:         iv,
			Ciphertext: ciphertext7,
		}
		rawResp, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(msg).
			Post(httpPrefix + acceptorAddr + api.Step7Endpoint)
		if err != nil {
			return ErrorMsg(err)
		}
		if rawResp.StatusCode() != http.StatusOK {
			return ErrorMsg(fmt.Errorf("step 7 status code is %d", rawResp.StatusCode()))
		}

		a.tui.session = acceptor
		a.tui.active[writeMessageItem] = struct{}{}

		return ErrorMsg(nil)
	}
}

func sendMessageCmd(a *Agent, msg string) tea.Cmd {
	return func() tea.Msg {
		if msg == "" {
			return ErrorMsg(nil)
		}

		iv, err := a.rng.GenerateIV()
		if err != nil {
			return ErrorMsg(err)
		}
		ciphertext := crypto.EncryptAES([]byte(msg), a.keys.sessionKey, iv)
		msg := api.Message{
			IV:         iv,
			Ciphertext: ciphertext,
		}
		rawResp, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(msg).
			Post(httpPrefix + a.cfg.AgentAddr + api.MessageEndpoint)
		if err != nil {
			return ErrorMsg(err)
		}
		if rawResp.StatusCode() != http.StatusOK {
			return ErrorMsg(fmt.Errorf("error sending message: status code is %d", rawResp.StatusCode()))
		}

		return ErrorMsg(nil)
	}
}

// Msg

type ModeChangedMsg int

type ErrorMsg error
