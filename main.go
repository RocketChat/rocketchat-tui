package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
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

type Model struct {
	textInput textinput.Model
	keys      *listKeyMap

	rlClient             *realtime.Client
	restClient           *rest.Client
	subscriptionList     []models.ChannelSubscription
	msgChannel           chan models.Message
	subscribed           map[string]string
	messageHistory       []models.Message
	activeChannel        models.ChannelSubscription
	lastMessageTimestamp *time.Time

	email    string
	password string
	token    string

	channelList          list.Model
	messagesList         list.Model
	slashCommandsList    list.Model
	slashCommands        []models.SlashCommand
	selectedSlashCommand *models.SlashCommand

	loginScreen *LoginScreen

	typing               bool
	loadMorePastMessages bool
	showSlashCommandList bool

	width  int
	height int
}

type listKeyMap struct {
	messageListNextPage             key.Binding
	messageListPreviousPage         key.Binding
	channelListNextChannel          key.Binding
	channelListPreviousChannel      key.Binding
	slashCommandListNextCommand     key.Binding
	slashCommandListPreviousCommand key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		messageListNextPage: key.NewBinding(
			key.WithKeys("ctrl+right"),
			key.WithHelp("ctrl+right", "Next Message Page"),
		),
		messageListPreviousPage: key.NewBinding(
			key.WithKeys("ctrl+left"),
			key.WithHelp("ctrl+left", "Previous Message Page"),
		),
		channelListNextChannel: key.NewBinding(
			key.WithKeys("ctrl+down"),
			key.WithHelp("ctrl+down", "Next Channel"),
		),
		channelListPreviousChannel: key.NewBinding(
			key.WithKeys("ctrl+up"),
			key.WithHelp("ctrl+up", "Previous Channel"),
		),
		slashCommandListNextCommand: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "Next Slash Command"),
		),
		slashCommandListPreviousCommand: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "Previous Slash Command"),
		),
	}
}

type LoginScreen struct {
	emailInput       textinput.Model
	passwordInput    textinput.Model
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

	e := textinput.NewModel()
	e.Placeholder = "Enter your email"
	e.Focus()

	p := textinput.NewModel()
	p.Placeholder = "Enter your password"
	p.Focus()

	intialLoginScreen := &LoginScreen{
		emailInput:       e,
		passwordInput:    p,
		activeElement:    1,
		loggedIn:         false,
		loginScreenState: "showLoginScreen",
		clickLoginButton: false,
	}

	t := textinput.NewModel()
	t.Placeholder = "Message"
	t.Focus()

	items := []list.Item{}
	listKeys := newListKeyMap()
	cl := list.New(items, channelListDelegate{}, w/4-1, 14)
	msgsList := list.New(items, messageListDelegate{}, 3*w/4-10, 16)
	slashCmndsList := list.New(items, slashCommandsListDelegate{}, 3*w/4-10, 5)

	cl.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.channelListNextChannel,
			listKeys.channelListPreviousChannel,
		}
	}

	msgsList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.messageListNextPage,
			listKeys.messageListPreviousPage,
		}
	}

	slashCmndsList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.slashCommandListNextCommand,
			listKeys.slashCommandListPreviousCommand,
		}
	}

	initialModel := &Model{
		channelList:          cl,
		keys:                 listKeys,
		messagesList:         msgsList,
		slashCommandsList:    slashCmndsList,
		loginScreen:          intialLoginScreen,
		textInput:            t,
		width:                w,
		height:               h,
		subscribed:           make(map[string]string),
		msgChannel:           make(chan models.Message, 100),
		loadMorePastMessages: false,
		showSlashCommandList: false,
		selectedSlashCommand: &models.SlashCommand{},
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
	err := CacheInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return tea.Batch(m.userLoginBegin(), m.waitForIncomingMessage(m.msgChannel))
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
				if m.loginScreen.activeElement < 3 {
					m.loginScreen.activeElement = m.loginScreen.activeElement + 1
				} else {
					m.loginScreen.activeElement = 1
				}
			case "enter":
				m.loginScreen.activeElement = 3
				if m.email != "" && m.password != "" {
					err := m.connectFromEmailAndPassword()
					if err != nil {
						os.Exit(1)
					}
					var cmds []tea.Cmd
					channelCmd := m.setChannelsInUiList()

					cmds = append(cmds, channelCmd, textinput.Blink)
					m.loginScreen.loggedIn = true
					m.changeSelectedChannel(0)
					return m, tea.Batch(cmds...)
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
		m.messagesList.Paginator.NextPage()
		m.loadMorePastMessages = false
		return m, tea.Batch(m.waitForIncomingMessage(m.msgChannel), cmd)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.channelListNextChannel):
			m.channelList.CursorDown()
			return m, nil
		case key.Matches(msg, m.keys.channelListPreviousChannel):
			m.channelList.CursorUp()
			return m, nil
		case key.Matches(msg, m.keys.messageListNextPage):
			m.messagesList.Paginator.NextPage()
			return m, nil
		case key.Matches(msg, m.keys.messageListPreviousPage):
			m.messagesList.Paginator.PrevPage()
			if m.messagesList.Paginator.Page == 0 && m.loadMorePastMessages {
				m.loadMorePastMessages = false
				msgsCmd := m.fetchPastMessages()
				return m, msgsCmd
			}
			if m.messagesList.Paginator.Page == 0 {
				m.loadMorePastMessages = true
			}
			return m, nil
		case key.Matches(msg, m.keys.slashCommandListNextCommand) && m.showSlashCommandList:
			m.slashCommandsList.CursorDown()
			return m, nil
		case key.Matches(msg, m.keys.slashCommandListPreviousCommand) && m.showSlashCommandList:
			m.slashCommandsList.CursorUp()
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if !m.typing {
				m.typing = true
				var msgItems []list.Item
				cmd := m.messagesList.SetItems(msgItems)
				m.changeSelectedChannel(m.channelList.Index())
				m.loadMorePastMessages = false
				return m, cmd
			}

			if m.typing {
				m, cmd := m.handleMessageInput()
				return m, cmd
			}
		case "esc":
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
		m, slashCmnd := m.handleShowingAndFilteringSlashCommandList()
		return m, tea.Batch(cmd, slashCmnd)
	}

	var channelCmd tea.Cmd
	m.channelList, channelCmd = m.channelList.Update(msg)

	var messageCmd tea.Cmd
	m.messagesList, messageCmd = m.messagesList.Update(msg)

	var slashCmd tea.Cmd
	m.slashCommandsList, slashCmd = m.slashCommandsList.Update(msg)

	cmds = append(cmds, channelCmd, messageCmd, slashCmd)
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
