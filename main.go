package main

import (
	// "encoding/json"
	"fmt"
	"strings"

	// "strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/rest"

	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var messageList []models.Message

type Model struct {
	textInput textinput.Model

	rlClient         *realtime.Client
	restClient       *rest.Client
	subscriptionList []models.ChannelSubscription
	msgChannel       chan models.Message
	subscribed       map[string]string
	messageHistory   []models.Message
	activeChannel    models.ChannelSubscription

	email    string
	password string
	token    string

	channelList  list.Model
	messagesList list.Model
	loginScreen  *LoginScreen

	updateMessageStreamCmd tea.Cmd

	loadChannels       bool
	typing             bool

	width  int
	height int
}

type LoginScreen struct {
	emailInput       textinput.Model
	passwordInput    textinput.Model
	authTokenInput   textinput.Model
	activeElement    int
	loginScreenState string
	loggedIn         bool
	clickLoginButton bool
}

func IntialModelState() *Model {
	w, h, err := term.GetSize(0)
	if err != nil {
		return nil
	}

	// credentials := getUserCredentails()

	e := textinput.NewModel()
	e.Placeholder = "Enter your email"
	e.Focus()

	p := textinput.NewModel()
	p.Placeholder = "Enter your password"
	p.Focus()

	at := textinput.NewModel()
	at.Placeholder = "Enter your auth token"
	at.Focus()

	intialLoginScreen := &LoginScreen{
		emailInput:       e,
		passwordInput:    p,
		authTokenInput:   at,
		activeElement:    1,
		loggedIn:         false,
		loginScreenState: "showLoginScreen",
		clickLoginButton: false,
	}

	t := textinput.NewModel()
	t.Placeholder = "Message"
	t.Focus()

	items := []list.Item{}
	l := list.New(items, channelListDelegate{}, w/4-1, 14)
	msgs := list.New(items, messageListDelegate{}, 3*w/4-10, 16)

	initialModel := &Model{
		channelList:            l,
		messagesList:           msgs,
		loginScreen:            intialLoginScreen,
		textInput:              t,
		width:                  w,
		height:                 h,
		subscribed:             make(map[string]string),
		msgChannel:             make(chan models.Message, 100),
		updateMessageStreamCmd: nil,
	}
	return initialModel
}

func main() {

	initialModel := IntialModelState()

	err := tea.NewProgram(initialModel, tea.WithAltScreen()).Start()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func (m *Model) waitForActivity(msgChannel chan models.Message) tea.Cmd {
	return func() tea.Msg {
		message := <-msgChannel
		if message.RoomID == m.activeChannel.RoomId {
			m.messageHistory = append(m.messageHistory, message)
			messageList = append(messageList, message)
			return message
		}
		return nil
	}
}

func (m *Model) Init() tea.Cmd {
	err := CacheInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return tea.Batch(m.userLoginBegin(), m.waitForActivity(m.msgChannel))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if !m.loginScreen.loggedIn && m.loginScreen.loginScreenState == "showLoginScreen" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "tab", "ctrl+down":
				if m.loginScreen.activeElement < 4 {
					m.loginScreen.activeElement = m.loginScreen.activeElement + 1
				} else {
					m.loginScreen.activeElement = 1
				}
			case "enter":
				if m.loginScreen.activeElement == 4 {
					fmt.Println("Login Using Auth Token")
				} else {
					m.loginScreen.activeElement = 3
					if m.email != "" && m.password != "" {
						// fmt.Println("Login User")
						err := m.connectFromEmailAndPassword()
						if err != nil {
							os.Exit(1)
						}
						// go m.handleMessageStream()

						var cmds []tea.Cmd
						channelCmd := m.setChannelsInUiList()

						cmds = append(cmds, channelCmd, textinput.Blink)
						m.loginScreen.loggedIn = true
						m.changeSelectedChannel(0)
						return m, tea.Batch(cmds...)
					}
				}
			}

		}

		if m.loginScreen.activeElement == 1 {
			var cmd tea.Cmd
			m.loginScreen.emailInput, cmd = m.loginScreen.emailInput.Update(msg)
			m.email = m.loginScreen.emailInput.Value()
			return m, cmd
		}

		if m.loginScreen.activeElement == 2 {
			var cmd tea.Cmd
			m.loginScreen.passwordInput, cmd = m.loginScreen.passwordInput.Update(msg)
			m.password = m.loginScreen.passwordInput.Value()
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case models.Message:
		msgItem := messagessItem(msg)
		cmd := m.messagesList.InsertItem(len(m.messagesList.Items()), msgItem)
		return m, tea.Batch(m.waitForActivity(m.msgChannel), cmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if !m.typing {
				m.typing = true
				var msgItems []list.Item
				cmd := m.messagesList.SetItems(msgItems)
				m.changeSelectedChannel(m.channelList.Index())
				return m, cmd
			}

			if m.typing {
				msg := strings.TrimSpace(m.textInput.Value())
				if msg != "" {
					m.sendMessage(msg)
					m.textInput.Reset()
					// PrintToLogFile(msg)
					return m, nil
				} else {
					m.textInput.Reset()
				}
			}

		case "ctrl+left":
			m.typing = false
			return m, nil

		case "ctrl+l", "ctrl+L":
			m, cmd := m.handleUserLogOut()
			return m, cmd

		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil
	}

	if m.typing {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	var channelCmd tea.Cmd
	m.channelList, channelCmd = m.channelList.Update(msg)

	var messageCmd tea.Cmd
	m.messagesList, messageCmd = m.messagesList.Update(msg)

	cmds = append(cmds, channelCmd, messageCmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {

	var completeUi string
	if m.loginScreen.loggedIn {
		completeUi = m.RenderTui()
	} else {
		completeUi = m.RenderLoginScreen()
	}

	dialog := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(completeUi),
	)
	return dialog
}
