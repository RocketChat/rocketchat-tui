package main

import (
	"fmt"
	"io"
	"strings"

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

type item models.ChannelSubscription

func (i item) FilterValue() string { return i.Name }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	// str := fmt.Sprintf("%d. %s", index+1, i)
	if i.Name != "" {
		nameLetter := sidebarTopColumnStyle.Align(lipgloss.Left).Render(nameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render(strings.ToUpper(string(i.Name[0:1]))))
		channelName := channelNameStyle.Copy().Bold(false).Render("# " + string(i.Name))

		if index == m.Index() {
			nameLetter = sidebarTopColumnStyle.Align(lipgloss.Left).Bold(true).Render(nameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render(strings.ToUpper(string(i.Name[0:1]))))
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

	channelList list.Model
	choice      string

	loadChannels bool

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
	l := list.New(items, itemDelegate{}, w/4-1, 14)

	initialModel := &Model{
		channelList: l,
		textInput:   t,
		width:       w,
		height:      h,
		email:       credentials.email,
		password:    credentials.pass,
	}
	err = tea.NewProgram(initialModel, tea.WithAltScreen()).Start()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func (m *Model) Init() tea.Cmd {
	go m.connect()
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			i, ok := m.channelList.SelectedItem().(item)
			if ok {
				m.choice = string(i.DisplayName)
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		return m, nil
	}

	if m.loadChannels {
		var items []list.Item
		for _, sub := range m.subscriptionList {
			if sub.Open && sub.Name != "" {
				items = append(items, item(sub))
			}
		}
		cmd := m.channelList.SetItems(items)
		m.loadChannels = false
		return m, cmd
	}

	var cmd tea.Cmd
	m.channelList, cmd = m.channelList.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	m.channelList.Title = "CHANNELS"
	m.channelList.SetShowStatusBar(false)
	m.channelList.SetFilteringEnabled(false)
	m.channelList.SetShowHelp(false)
	m.channelList.Styles.Title = titleStyle.Width(m.width/4 - 1)
	m.channelList.Styles.PaginationStyle = paginationStyle
	m.channelList.Styles.HelpStyle = helpStyle

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
	starIcon := starIconStyle.Render("☆")

	channelWindowTitle := channelWindowTitleStyle.Width((3 * m.width / 8) - 2).Render(lipgloss.JoinHorizontal(lipgloss.Top, channelTopbarNameLetterBox, channelName, starIcon))
	channelOptionsButton := channelOptionsButtonStyle.Width((3 * m.width / 8)).Render(nameLetterBoxStyle.Background(lipgloss.Color("#13505b")).Render("⠇"))
	channelWindowTopbar := channelWindowTopbarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, channelWindowTitle, channelOptionsButton))

	channelConversationScreen := lipgloss.NewStyle().Height(m.height - 7).Render("Here comes messages")

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
