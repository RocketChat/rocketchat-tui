package main

import (
	// "encoding/json"
	"fmt"
	"io"
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

type channelsItem models.ChannelSubscription

func (i channelsItem) FilterValue() string { return i.Name }

type messagessItem models.Message

func (i messagessItem) FilterValue() string { return i.Timestamp.String() }

type messageListDelegate struct{}

func (d messageListDelegate) Height() int  { return 1 }
func (d messageListDelegate) Spacing() int { return 0 }
func (d messageListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d messageListDelegate) Render(w io.Writer, m list.Model, index int, messageListItem list.Item) {
	i, ok := messageListItem.(messagessItem)
	if !ok {
		return
	}
	nameLetterChat := nameLetterBoxStyle.Copy().Width(3).Render(getStringFirstLetter(i.User.Name))

	userFullName := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).PaddingRight(1).Bold(true).Align(lipgloss.Left).Render("Sitaram Rathi")
	userName := lipgloss.NewStyle().Foreground(lipgloss.Color("#767373")).PaddingRight(1).Bold(true).Align(lipgloss.Left).Render("@" + i.User.UserName)
	timeStamp := lipgloss.NewStyle().Foreground(lipgloss.Color("#767373")).Align(lipgloss.Left).Render(i.Timestamp.Format("15:04"))
	userDetails := lipgloss.JoinHorizontal(lipgloss.Left, userFullName, userName, timeStamp)

	userMessage := lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#ffffff")).MaxWidth(80).Width(80).Render(i.Msg)
	messageBox := lipgloss.NewStyle().PaddingLeft(1).Render(lipgloss.JoinVertical(lipgloss.Top, userDetails, userMessage))
	userMessageBox := lipgloss.NewStyle().PaddingBottom(1).Width(80).Render(lipgloss.JoinHorizontal(lipgloss.Left, nameLetterChat, messageBox))
	// messageString = userMessageBox
	// messageString += userMessageBox + "\n"
	// fmt.Println(userMessageBox)

	fmt.Fprintf(w, userMessageBox)

	// str := fmt.Sprintf("%d. %s", index+1, i)
}

type channelListDelegate struct{}

func (d channelListDelegate) Height() int  { return 1 }
func (d channelListDelegate) Spacing() int { return 0 }
func (d channelListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d channelListDelegate) Render(w io.Writer, m list.Model, index int, channelListItem list.Item) {
	i, ok := channelListItem.(channelsItem)
	if !ok {
		return
	}
	if i.Name != "" {
		nameLetter := sidebarTopColumnStyle.Align(lipgloss.Left).Render(nameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render(getStringFirstLetter(i.Name)))
		channelName := channelNameStyle.Copy().Bold(false).Render("# " + string(i.Name))

		if index == m.Index() {
			nameLetter = sidebarTopColumnStyle.Align(lipgloss.Left).Bold(true).Render(nameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render(getStringFirstLetter(i.Name)))
			channelName = channelNameStyle.Copy().Bold(true).Underline(true).Render("# " + string(i.Name))
		}
		str := channelWindowTitleStyle.Copy().PaddingBottom(1).Render(lipgloss.JoinHorizontal(lipgloss.Top, nameLetter, channelName))

		fmt.Fprintf(w, str)
	}
}

type Model struct {
	textInput textinput.Model

	rlClient         *realtime.Client
	restClient       *rest.Client
	subscriptionList []models.ChannelSubscription
	msgChannel       chan models.Message
	subscribed       map[string]string
	messageHistory   []models.Message
	activeChannel    models.ChannelSubscription
	email            string
	password         string

	channelList  list.Model
	messagesList list.Model
	choice       string
	loginScreen  LoginScreen

	updateMessageStream    bool
	updateMessageStreamCmd tea.Cmd

	loadChannels bool
	typing       bool

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
		loginScreenState: "emailTyping",
		loggedIn:         false,
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
		loginScreen:            *intialLoginScreen,
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

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if !m.loginScreen.loggedIn {
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
						err := m.connect()
						if err != nil {
							os.Exit(1)
						}
						go m.handleMessageStream()

						var cmds []tea.Cmd
						channelCmd := m.setChannelsInUiList()

						cmds = append(cmds, channelCmd, textinput.Blink)
						m.loginScreen.loggedIn = true
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

	// PrintToLogFile("MESSAGES STREAM", messageList)
	// PrintToLogFile("MESSAGE LIST", m.messagesList.Items())

	if m.updateMessageStreamCmd != nil {
		cmd := m.updateMessageStreamCmd
		m.updateMessageStreamCmd = nil
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if !m.typing {
				m.typing = true
				m, messagesCmd := m.changeAndPopulateChannelMessages()
				// bs, _ := json.Marshal(messageList)
				// PrintToLogFile(string(bs))
				return m, messagesCmd
			}

			if m.typing {
				msg := strings.TrimSpace(m.textInput.Value())
				if msg != "" {
					m.sendMessage(msg)
					m.textInput.Reset()

					// m, messagesCmd := m.changeAndPopulateChannelMessages()
					// bs, _ := json.Marshal(messageList)
					// PrintToLogFile(string(bs))
					return m, nil
				} else {
					m.textInput.Reset()
				}
			}

		case "ctrl+left":
			m.typing = false
			return m, nil

		case "ctrl+l", "ctrl+L":
			m.handleUserLogOut()

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

	// if m.loadMessages {
	// 	m.loadMessages = false
	// 	var messageCmd tea.Cmd
	// 	m.messagesList, messageCmd = m.messagesList.Update(msg)
	// 	return m, messageCmd
	// }

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
