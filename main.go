package main

import (
	"encoding/json"
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

	loadChannels bool
	typing       bool

	width  int
	height int
}

func main() {
	w, h, err := term.GetSize(0)
	if err != nil {
		return
	}

	credentials := getUserCredentails()

	t := textinput.NewModel()
	t.Placeholder = "Message"
	t.Focus()

	items := []list.Item{}
	l := list.New(items, channelListDelegate{}, w/4-1, 14)
	msgs := list.New(items, messageListDelegate{}, 3*w/4-10, 16)

	initialModel := &Model{
		channelList:  l,
		messagesList: msgs,
		textInput:    t,
		width:        w,
		height:       h,
		subscribed:   make(map[string]string),
		email:        credentials.email,
		password:     credentials.pass,
		msgChannel:   make(chan models.Message, 100),
	}
	err = tea.NewProgram(initialModel, tea.WithAltScreen()).Start()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func (m *Model) Init() tea.Cmd {
	m.connect()
	go m.handleMessageStream()

	var cmds []tea.Cmd
	channelCmd := m.setChannelsInUiList()

	cmds = append(cmds, channelCmd, textinput.Blink)
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if !m.typing {
				m.typing = true
				m, messagesCmd := m.changeAndPopulateChannelMessages()
				bs, _ := json.Marshal(messageList)
				PrintToLogFile(string(bs))
				return m, messagesCmd
			}

			if m.typing {
				msg := strings.TrimSpace(m.textInput.Value())
				if msg != "" {
					m.sendMessage(msg)
					m.textInput.Reset()

					m, messagesCmd := m.changeAndPopulateChannelMessages()
					bs, _ := json.Marshal(messageList)
					PrintToLogFile(string(bs))
					return m, messagesCmd
				} else {
					m.textInput.Reset()
				}
			}

		case "ctrl+left":
			m.typing = false
			return m, nil
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
	m.channelList.Title = "CHANNELS"
	m.channelList.SetShowStatusBar(false)
	m.channelList.SetFilteringEnabled(false)
	m.channelList.SetShowHelp(false)
	m.channelList.Styles.Title = titleStyle.Width(m.width/4 - 1)
	m.channelList.Styles.PaginationStyle = paginationStyle
	m.channelList.Styles.HelpStyle = helpStyle

	m.messagesList.SetShowTitle(false)
	m.messagesList.SetShowStatusBar(false)
	m.messagesList.SetFilteringEnabled(false)
	m.messagesList.SetShowHelp(false)
	m.messagesList.SetShowPagination(false)

	nameLetter := sidebarTopColumnStyle.Width(m.width / 8).Align(lipgloss.Left).Render(nameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render("S"))
	newChannelButton := sidebarTopColumnStyle.Width(m.width / 8).Align(lipgloss.Right).Render(nameLetterBoxStyle.Background(lipgloss.Color("#13505b")).Render("✍."))

	sidebarTopbar := sidebarTopbarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, nameLetter, newChannelButton))
	sidebarListBox := lipgloss.NewStyle().Height(m.height - 5).Render(m.channelList.View())
	rocketChatIcon := rocketChatIconStyle.Render("rocket.chat")

	sidebar := lipgloss.Place(m.width/4, m.height-2,
		lipgloss.Left, lipgloss.Top,
		sidebarStyle.Height(m.height-2).Width(m.width/4).Render(lipgloss.JoinVertical(lipgloss.Top, sidebarTopbar, sidebarListBox, rocketChatIcon)),
	)

	channelTopbarNameLetterBox := nameLetterBoxStyle.Background(lipgloss.Color("#edae49")).Bold(true).Render("R")
	channelName := channelNameStyle.Render("# Rocket.Chat Terminal TUI")

	if m.activeChannel.Name != "" && m.activeChannel.Open {
		channelTopbarNameLetterBox = nameLetterBoxStyle.Background(lipgloss.Color("#edae49")).Bold(true).Render(getStringFirstLetter(m.activeChannel.Name))
		channelName = channelNameStyle.Render("# " + m.activeChannel.Name)
	}
	starIcon := starIconStyle.Render("☆")

	channelWindowTitle := channelWindowTitleStyle.Width((3 * m.width / 8) - 2).Render(lipgloss.JoinHorizontal(lipgloss.Top, channelTopbarNameLetterBox, channelName, starIcon))
	channelOptionsButton := channelOptionsButtonStyle.Width((3 * m.width / 8)).Render(nameLetterBoxStyle.Background(lipgloss.Color("#13505b")).Render("⠇"))
	channelWindowTopbar := channelWindowTopbarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, channelWindowTitle, channelOptionsButton))

	channelConversationScreen := lipgloss.NewStyle().Height(m.height - 7).Render(m.messagesList.View())
	messageEmojiIcon := messageEmojiIconStyle.Render("☺")
	channelMessageInputBox := channelMessageInputBoxStyle.Width((3 * m.width / 4) - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, messageEmojiIcon, m.textInput.View()))

	channelWindow := lipgloss.Place(3*(m.width/4)-2, m.height-2,
		lipgloss.Left, lipgloss.Top,
		channelWindowStyle.Height(m.height-2).Width(3*(m.width/4)).Render(lipgloss.JoinVertical(lipgloss.Top, channelWindowTopbar, channelConversationScreen, channelMessageInputBox)),
	)

	instruction := instructionStyle.Width(m.width).Render("Press Ctrl + C - quit • Ctrl + H - help • Arrows - for navigation in pane • Enter - send message")
	ui := lipgloss.JoinHorizontal(lipgloss.Center, sidebar, channelWindow)
	completeUi := lipgloss.JoinVertical(lipgloss.Center, instruction, ui)
	dialog := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(completeUi),
	)
	return dialog
}
