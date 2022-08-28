package ui

import (
	"fmt"
	"io"
	"strconv"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/rocketchat-tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ListDelegate interface {
	Height() int
	Spacing() int
	Update(msg tea.Msg, m *list.Model) tea.Cmd
	Render(w io.Writer, m list.Model, index int, messageListItem list.Item)
}

type MessagessItem models.Message

func (i MessagessItem) FilterValue() string { return i.Timestamp.String() }

type MessageListDelegate struct{}

func (d MessageListDelegate) Height() int  { return 1 }
func (d MessageListDelegate) Spacing() int { return 0 }
func (d MessageListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d MessageListDelegate) Render(w io.Writer, m list.Model, index int, messageListItem list.Item) {
	i, ok := messageListItem.(MessagessItem)
	if !ok {
		return
	}
	nameLetterChat := styles.NameLetterBoxStyle.Copy().Width(3).Render(getStringFirstLetter(i.User.Name))

	userFullName := styles.UserFullNameStyle.Render(i.User.Name)
	userName := styles.UsernameStyle.Render("@" + i.User.UserName)
	timeStamp := styles.TimestampStyle.Render(i.Timestamp.Format("15:04"))
	userDetails := lipgloss.JoinHorizontal(lipgloss.Left, userFullName, userName, timeStamp)

	userMessage := styles.UserMessageStyle.Render(i.Msg)
	messageBox := styles.MessageBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Top, userDetails, userMessage))

	userMessageBox := styles.UserMessageBoxStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, nameLetterChat, messageBox))

	fmt.Fprintf(w, userMessageBox)
}

type ChannelsItem models.ChannelSubscription

func (i ChannelsItem) FilterValue() string { return i.Name }

type ChannelListDelegate struct{}

func (d ChannelListDelegate) Height() int  { return 1 }
func (d ChannelListDelegate) Spacing() int { return 1 }
func (d ChannelListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d ChannelListDelegate) Render(w io.Writer, m list.Model, index int, channelListItem list.Item) {
	i, ok := channelListItem.(ChannelsItem)
	if !ok {
		return
	}
	if i.Name != "" {
		nameLetter := styles.SidebarTopColumnStyle.Align(lipgloss.Left).Render(styles.NameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render(getStringFirstLetter(i.Name)))
		channelName := styles.ChannelNameStyle.Copy().Bold(false).Render("# " + string(i.Name))

		if index == m.Index() {
			nameLetter = styles.SidebarTopColumnStyle.Align(lipgloss.Left).Bold(true).Render(styles.NameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render(getStringFirstLetter(i.Name)))
			channelName = styles.ChannelNameStyle.Copy().Bold(true).Underline(true).Render("# " + string(i.Name))
		}
		str := styles.ChannelWindowTitleStyle.Copy().PaddingBottom(1).Render(lipgloss.JoinHorizontal(lipgloss.Top, nameLetter, channelName))

		fmt.Fprintf(w, str)
	}
}

type SlashCommandsItem models.SlashCommand

func (i SlashCommandsItem) FilterValue() string { return i.Command }

type SlashCommandsListDelegate struct{}

func (d SlashCommandsListDelegate) Height() int  { return 1 }
func (d SlashCommandsListDelegate) Spacing() int { return 0 }
func (d SlashCommandsListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d SlashCommandsListDelegate) Render(w io.Writer, m list.Model, index int, slashCommandsListItem list.Item) {
	i, ok := slashCommandsListItem.(SlashCommandsItem)
	if !ok {
		return
	}
	if i.Command != "" {
		slashCommand := styles.SlashCommandStyle.Copy().Foreground(lipgloss.Color("ffffff")).Bold(false).MaxWidth(120).Render("/" + string(i.Command))
		if index == m.Index() {
			slashCommand = styles.SlashCommandStyle.Render("/" + string(i.Command))
		}
		fmt.Fprintf(w, slashCommand)
	}

}

type ChannelMembersItem models.User

func (i ChannelMembersItem) FilterValue() string { return i.UserName }

type ChannelMembersListDelegate struct{}

func (c ChannelMembersListDelegate) Height() int  { return 1 }
func (c ChannelMembersListDelegate) Spacing() int { return 0 }
func (c ChannelMembersListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (c ChannelMembersListDelegate) Render(w io.Writer, m list.Model, index int, channelMembersListItem list.Item) {
	i, ok := channelMembersListItem.(ChannelMembersItem)
	var activeStatus string
	if i.Status == "online" {
		activeStatus = "●"
	} else if i.Status == "offline" {
		activeStatus = "○"
	} else {
		activeStatus = "◦"
	}
	if !ok {
		return
	}
	if i.UserName != "" {
		var nameLetterChat string
		if i.Status != "" {
			nameLetterChat = styles.NameLetterBoxStyle.Copy().MarginLeft(1).Width(3).Render(getStringFirstLetter(i.Name))
		}
		userName := styles.ChannelMembersListUsernameStyle.Render(i.UserName)
		name := styles.ChannelMembersListNameStyle.Render(i.Name)
		channelMemberItem := styles.ChannelNameStyle.Copy().Bold(false).Align(lipgloss.Left).UnsetPaddingTop().UnsetMarginTop().MaxWidth(120).Render(activeStatus + "   " + nameLetterChat + " " + userName + " " + name)
		if index == m.Index() {
			userName := styles.ChannelMembersListUsernameStyle.Copy().Foreground(lipgloss.Color("#119da4")).Render(i.UserName)
			name := styles.ChannelMembersListNameStyle.Copy().Foreground(lipgloss.Color("#119da4")).Render(i.Name)
			channelMemberItem = styles.ChannelNameStyle.Copy().Align(lipgloss.Left).UnsetPaddingTop().UnsetMarginTop().Foreground(lipgloss.Color("#119da4")).Render(activeStatus + "   " + nameLetterChat + " " + userName + " " + name)
		}
		fmt.Fprintf(w, channelMemberItem)
	}

}

func (m *Model) RenderTui() string {
	cmndboxSpace := 8
	if m.showSlashCommandList {
		cmndboxSpace = 13
	} else if m.showChannelMembersList {
		cmndboxSpace = 13
	}

	m.channelList.Title = "CHANNELS"
	m.channelList.SetShowStatusBar(false)
	m.channelList.SetFilteringEnabled(false)
	m.channelList.SetShowHelp(false)
	m.channelList.Styles.Title = styles.TitleStyle.Width(m.width/4 - 1)
	m.channelList.Styles.PaginationStyle = styles.PaginationStyle
	m.channelList.Styles.HelpStyle = styles.HelpStyle

	m.messagesList.SetShowTitle(false)
	m.messagesList.SetShowStatusBar(false)
	m.messagesList.SetFilteringEnabled(false)
	m.messagesList.SetShowHelp(false)
	m.messagesList.SetShowPagination(false)

	m.slashCommandsList.Title = "COMMANDS"
	m.slashCommandsList.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).UnsetMargins().UnsetPadding()
	m.slashCommandsList.SetShowFilter(true)
	m.slashCommandsList.SetFilteringEnabled(true)
	m.slashCommandsList.SetShowStatusBar(false)
	m.slashCommandsList.SetShowHelp(false)
	m.slashCommandsList.SetShowPagination(false)

	m.channelMembersList.Title = "MEMBERS"
	m.channelMembersList.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).UnsetMargins().UnsetPadding()
	m.channelMembersList.SetShowFilter(true)
	m.channelMembersList.SetFilteringEnabled(true)
	m.channelMembersList.SetShowStatusBar(false)
	m.channelMembersList.SetShowHelp(false)
	m.channelMembersList.SetShowPagination(false)

	nameLetter := styles.SidebarTopColumnStyle.Width(m.width / 8).Align(lipgloss.Left).Render(styles.NameLetterBoxStyle.Background(lipgloss.Color("#d1495b")).Bold(true).Render("S"))
	newChannelButton := styles.SidebarTopColumnStyle.Width(m.width / 8).Align(lipgloss.Right).Render(styles.NameLetterBoxStyle.Background(lipgloss.Color("#13505b")).Render("✍."))

	sidebarTopbar := styles.SidebarTopbarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, nameLetter, newChannelButton))
	sidebarListBox := lipgloss.NewStyle().Height(m.height - 5).Render(m.channelList.View())
	rocketChatIcon := styles.RocketChatIconStyle.Render("rocket.chat")

	sidebar := lipgloss.Place(m.width/4, m.height-2,
		lipgloss.Left, lipgloss.Top,
		styles.SidebarStyle.Height(m.height-2).Width(m.width/4).Render(lipgloss.JoinVertical(lipgloss.Top, sidebarTopbar, sidebarListBox, rocketChatIcon)),
	)

	channelTopbarNameLetterBox := styles.NameLetterBoxStyle.Background(lipgloss.Color("#edae49")).Bold(true).Render("R")
	channelName := styles.ChannelNameStyle.Render("# Rocket.Chat Terminal TUI")

	if m.activeChannel.Name != "" && m.activeChannel.Open {
		channelTopbarNameLetterBox = styles.NameLetterBoxStyle.Background(lipgloss.Color("#edae49")).Bold(true).Render(getStringFirstLetter(m.activeChannel.Name))
		channelName = styles.ChannelNameStyle.Render("# " + m.activeChannel.Name)
	}
	starIcon := styles.StarIconStyle.Render("☆")

	channelWindowTitle := styles.ChannelWindowTitleStyle.Width((3 * m.width / 8) - 2).Render(lipgloss.JoinHorizontal(lipgloss.Top, channelTopbarNameLetterBox, channelName, starIcon))
	messagePagesCount := styles.ChannelOptionsButtonStyle.Width((3 * m.width / 8)).Render(styles.NameLetterBoxStyle.Background(lipgloss.Color("#13505b")).Render(strconv.Itoa(m.messagesList.Paginator.Page+1) + "/" + strconv.Itoa((m.messagesList.Paginator.TotalPages))))
	channelWindowTopbar := styles.ChannelWindowTopbarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, channelWindowTitle, messagePagesCount))

	channelConversationScreen := lipgloss.NewStyle().Height(m.height - cmndboxSpace).MaxHeight(m.height - cmndboxSpace).Render(m.messagesList.View())

	var smallListBox string
	if m.showSlashCommandList {
		smallListBox = styles.SmallListBoxStyle.Width((3 * m.width / 4) - 4).Render(m.slashCommandsList.View())
	} else if m.showChannelMembersList {
		smallListBox = styles.SmallListBoxStyle.Width((3 * m.width / 4) - 4).Render(m.channelMembersList.View())
	}

	messageEmojiIcon := styles.MessageEmojiIconStyle.Render("☺")

	var channelMessageInputBox string
	if m.typing {
		channelMessageInputBox = styles.ChannelMessageInputBoxStyle.Copy().BorderForeground(lipgloss.Color("#ffffff")).Width((3 * m.width / 4) - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, messageEmojiIcon, m.textInput.View()))
	} else {
		channelMessageInputBox = styles.ChannelMessageInputBoxStyle.Width((3 * m.width / 4) - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, messageEmojiIcon, m.textInput.View()))
	}

	channelWindow := lipgloss.Place(3*(m.width/4)-2, m.height-2,
		lipgloss.Left, lipgloss.Top,
		styles.ChannelWindowStyle.Height(m.height-2).Width(3*(m.width/4)).Render(lipgloss.JoinVertical(lipgloss.Top, channelWindowTopbar, channelConversationScreen, smallListBox, channelMessageInputBox)),
	)

	instruction := styles.InstructionStyle.Width(m.width).Render("Press Ctrl + C - quit • Ctrl + H - help • Ctrl + Arrows - for navigation in pane • Enter - send message • Ctrl + L - Log out")
	ui := lipgloss.JoinHorizontal(lipgloss.Center, sidebar, channelWindow)
	completeUi := lipgloss.JoinVertical(lipgloss.Center, instruction, ui)
	return completeUi
}

func (m *Model) RenderLoginScreen() string {
	loginScreenWelcomeText := styles.LoginScreenWelcomeTextStyle.Render("WELCOME TO ROCKET.CHAT")
	loginHeadingText := styles.LoginHeadingTextStyle.Render("Login into your Account")

	emailInputLabel := styles.EmailInputLabelStyle.Render("E-mail")
	var emailInputBox string
	if m.loginScreen.activeElement == 1 {
		emailInputBox = styles.LoginScreenInputBoxActiveStyle.Render(m.loginScreen.emailInput.View())
	} else {
		emailInputBox = styles.LoginScreenInputBoxNotActiveStyle.Render(m.loginScreen.emailInput.View())
	}

	passowrdInputLabel := styles.PasswordInputLabelStyle.Render("Password")
	var passowrdInputBox string
	if m.loginScreen.activeElement == 2 {
		passowrdInputBox = styles.LoginScreenInputBoxActiveStyle.Render(m.loginScreen.passwordInput.View())
	} else {
		passowrdInputBox = styles.LoginScreenInputBoxNotActiveStyle.Render(m.loginScreen.passwordInput.View())
	}

	var loginButton string
	if m.loginScreen.activeElement == 3 {
		loginButton = styles.LoginButtonActiveStyle.Render("LOG INTO TUI")
	} else {
		loginButton = styles.LoginButtonNotActiveStyle.Render("LOG INTO TUI")
	}

	loginUiBox := styles.LoginUiBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Top, loginScreenWelcomeText, loginHeadingText, emailInputLabel, emailInputBox, passowrdInputLabel, passowrdInputBox, loginButton))

	return loginUiBox
}
