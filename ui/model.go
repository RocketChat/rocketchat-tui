package ui

import (
	"log"
	"os"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/rest"
	"golang.org/x/term"

	"github.com/RocketChat/rocketchat-tui/cache"
	"github.com/RocketChat/rocketchat-tui/keyBindings"
	"github.com/RocketChat/rocketchat-tui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// All Bubble tea required core functions like Init, Update and View will be method of this.
// It defines the Global State of the TUI.
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

	email     string
	password  string
	token     string
	serverUrl string

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

// It cantains the state for the Loginscreen.
type LoginScreen struct {
	emailInput       textinput.Model
	passwordInput    textinput.Model
	activeElement    int
	loginScreenState string
	loggedIn         bool
	clickLoginButton bool
	err              error
}

// It will generate the initial global model state for the TUI.
func IntialModelState(sUrl string) *Model {
	// To get the width and height of initial terminal screen.
	w, h, err := term.GetSize(0)
	if err != nil {
		return nil
	}

	// Email text input component for Loginscreen.
	e := textinput.NewModel()
	e.Placeholder = "Enter your email"
	e.Width = 42
	e.Focus()

	// Password text input component for Loginscreen.. In Echomode we can hide the text typing.
	p := textinput.NewModel()
	p.Placeholder = "Enter your password"
	p.Width = 42
	p.EchoMode = textinput.EchoPassword
	p.EchoCharacter = '•'
	p.Focus()

	// Inital login scrfeen state.
	intialLoginScreen := &LoginScreen{
		emailInput:       e,
		passwordInput:    p,
		activeElement:    1,
		loggedIn:         false,
		loginScreenState: "showLoginScreen",
		clickLoginButton: false,
		err:              nil,
	}

	// Message text input component for TUI
	t := textinput.NewModel()
	t.Placeholder = "Message"
	t.Focus()

	// List component of bubble tea for holding channels list.
	items := []list.Item{}
	listKeys := keyBindings.NewListKeyMap()

	channelListDelegate := ChannelListDelegate{}
	msgListDelegate := MessageListDelegate{}
	slashCmndsListDelegate := SlashCommandsListDelegate{}
	channelMembersListDelegate := ChannelMembersListDelegate{}

	// All lists will contain Height, Spacing, Render and Update function. All lists will be of type 'ListDelegate'
	listDelegates := []ListDelegate{channelListDelegate, msgListDelegate, slashCmndsListDelegate, channelMembersListDelegate}

	cl := list.New(items, listDelegates[0], w/4-1, 14)
	msgsList := list.New(items, listDelegates[1], 3*w/4-10, 16)
	slashCmndsList := list.New(items, listDelegates[2], 3*w/4-10, 5)
	channelMembersList := list.New(items, listDelegates[3], 3*w/4-10, 5)

	// Adding key bindings to the lists.
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

	// Initial global model state.
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
		serverUrl:              sUrl,
	}
	return initialModel
}

// This the function which will initialise the TUI. All init functions should be called in it.
func (m *Model) Init() tea.Cmd {
	err := cache.CacheInit()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return tea.Batch(m.userLoginBegin(), m.waitForIncomingMessage(m.msgChannel))
}

// This is the Main Update function which updates the TUI.
// All other update functions to update state and TUI are called in it.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var channelCmd, messageCmd, slashCmd, channelMembersListCmnd tea.Cmd

	loginScreenUpdateCmd, err := m.handleLoginScreenUpdate(msg)
	if loginScreenUpdateCmd != nil {
		return m, loginScreenUpdateCmd
	} else if err != nil {
		m.loginScreen.err = err
		return m, nil
	}

	switch msg := msg.(type) {
	case models.Message:
		msgItem := MessagessItem(msg)
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
		m.textInput.Width = (3 * m.width / 4) - 10
		return m, nil
	}

	m.channelList, channelCmd = m.channelList.Update(msg)
	m.messagesList, messageCmd = m.messagesList.Update(msg)
	m.slashCommandsList, slashCmd = m.slashCommandsList.Update(msg)
	m.channelMembersList, channelMembersListCmnd = m.channelMembersList.Update(msg)

	cmds = append(cmds, channelCmd, messageCmd, slashCmd, channelMembersListCmnd)
	return m, tea.Batch(cmds...)
}

// This renders the UI of the TUI.
func (m *Model) View() string {

	var completeUi string
	if m.loginScreen.loggedIn {
		completeUi = m.RenderTui()
	} else {
		completeUi = m.RenderLoginScreen()
	}
	dialog := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.DialogStyle.Render(completeUi),
	)
	return dialog
}
