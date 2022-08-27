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
	"github.com/RocketChat/rocketchat-tui/keyBindings"

	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Model struct {
	textInput textinput.Model
	keys      *keyBindings.ListKeyMap

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

	channelList        list.Model
	messagesList       list.Model
	slashCommandsList  list.Model
	channelMembersList list.Model

	channelMembers       []models.User
	slashCommands        []models.SlashCommand
	selectedSlashCommand *models.SlashCommand

	loginScreen *LoginScreen

	typing                 bool
	loadMorePastMessages   bool
	showSlashCommandList   bool
	showChannelMembersList bool

	positionOfAtSymbol int
	width              int
	height             int
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
	p.EchoMode = textinput.EchoPassword
	p.EchoCharacter = '•'
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
	listKeys := keyBindings.NewListKeyMap()
	cl := list.New(items, channelListDelegate{}, w/4-1, 14)
	msgsList := list.New(items, messageListDelegate{}, 3*w/4-10, 16)
	slashCmndsList := list.New(items, slashCommandsListDelegate{}, 3*w/4-10, 5)
	channelMembersList := list.New(items, channelMembersListDelegate{}, 3*w/4-10, 5)

	cl.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ChannelListNextChannel,
			listKeys.ChannelListPreviousChannel,
		}
	}

	msgsList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.MessageListNextPage,
			listKeys.MessageListPreviousPage,
		}
	}

	slashCmndsList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.SlashCommandListNextCommand,
			listKeys.SlashCommandListPreviousCommand,
		}
	}

	channelMembersList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.ChannelMembersListNextMember,
			listKeys.ChannelMembersListPreviousMember,
		}
	}

	initialModel := &Model{
		channelList:            cl,
		keys:                   listKeys,
		messagesList:           msgsList,
		slashCommandsList:      slashCmndsList,
		loginScreen:            intialLoginScreen,
		textInput:              t,
		width:                  w,
		height:                 h,
		subscribed:             make(map[string]string),
		msgChannel:             make(chan models.Message, 100),
		loadMorePastMessages:   false,
		showSlashCommandList:   false,
		selectedSlashCommand:   &models.SlashCommand{},
		channelMembersList:     channelMembersList,
		showChannelMembersList: false,
		positionOfAtSymbol:     -1,
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
	var channelCmd, messageCmd, slashCmd, channelMembersListCmnd tea.Cmd

	loginScreenUpdateCmd := m.handleLoginScreenUpdate(msg)
	if loginScreenUpdateCmd != nil {
		return m, loginScreenUpdateCmd
	}

	switch msg := msg.(type) {
	case models.Message:
		msgItem := messagessItem(msg)
		cmd := m.messagesList.InsertItem(len(m.messagesList.Items()), msgItem)
		m.messagesList.Paginator.NextPage()
		m.loadMorePastMessages = false
		return m, tea.Batch(m.waitForIncomingMessage(m.msgChannel), cmd)

	case tea.KeyMsg:
		m, cmd := m.handleUpdateOnKeyPress(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil
	}

	m.channelList, channelCmd = m.channelList.Update(msg)

	m.messagesList, messageCmd = m.messagesList.Update(msg)

	m.slashCommandsList, slashCmd = m.slashCommandsList.Update(msg)

	m.channelMembersList, channelMembersListCmnd = m.channelMembersList.Update(msg)

	cmds = append(cmds, channelCmd, messageCmd, slashCmd, channelMembersListCmnd)
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
