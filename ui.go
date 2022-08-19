package main

import (
	"fmt"
	"io"
	"strconv"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

	userFullName := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).PaddingRight(1).Bold(true).Align(lipgloss.Left).Render(i.User.Name)
	userName := lipgloss.NewStyle().Foreground(lipgloss.Color("#767373")).PaddingRight(1).Bold(true).Align(lipgloss.Left).Render("@" + i.User.UserName)
	timeStamp := lipgloss.NewStyle().Foreground(lipgloss.Color("#767373")).Align(lipgloss.Left).Render(i.Timestamp.Format("15:04"))
	userDetails := lipgloss.JoinHorizontal(lipgloss.Left, userFullName, userName, timeStamp)

	userMessage := lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#ffffff")).MaxWidth(80).Width(80).Render(i.Msg)
	messageBox := lipgloss.NewStyle().PaddingLeft(1).Render(lipgloss.JoinVertical(lipgloss.Top, userDetails, userMessage))

	userMessageBox := lipgloss.NewStyle().PaddingBottom(1).Width(80).Render(lipgloss.JoinHorizontal(lipgloss.Left, nameLetterChat, messageBox))

	fmt.Fprintf(w, userMessageBox)
}

type channelsItem models.ChannelSubscription

func (i channelsItem) FilterValue() string { return i.Name }

type channelListDelegate struct{}

func (d channelListDelegate) Height() int  { return 1 }
func (d channelListDelegate) Spacing() int { return 1 }
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

type slashCommandsItem models.SlashCommand

func (i slashCommandsItem) FilterValue() string { return i.Command }

type slashCommandsListDelegate struct{}

func (d slashCommandsListDelegate) Height() int  { return 1 }
func (d slashCommandsListDelegate) Spacing() int { return 0 }
func (d slashCommandsListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d slashCommandsListDelegate) Render(w io.Writer, m list.Model, index int, slashCommandsListItem list.Item) {
	i, ok := slashCommandsListItem.(slashCommandsItem)
	if !ok {
		return
	}
	if i.Command != "" {
		slashCommand := channelNameStyle.Copy().Bold(false).Align(lipgloss.Left).UnsetPaddingTop().UnsetMarginTop().MaxWidth(120).Render("/" + string(i.Command))
		if index == m.Index() {
			slashCommand = channelNameStyle.Copy().Align(lipgloss.Left).UnsetPaddingTop().UnsetMarginTop().Foreground(lipgloss.Color("#119da4")).Render("/" + string(i.Command))
		}
		fmt.Fprintf(w, slashCommand)
	}

}

type channelMembersItem models.User

func (i channelMembersItem) FilterValue() string { return i.UserName }

type channelMembersListDelegate struct{}

func (c channelMembersListDelegate) Height() int  { return 1 }
func (c channelMembersListDelegate) Spacing() int { return 0 }
func (c channelMembersListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (c channelMembersListDelegate) Render(w io.Writer, m list.Model, index int, channelMembersListItem list.Item) {
	i, ok := channelMembersListItem.(channelMembersItem)
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
			nameLetterChat = nameLetterBoxStyle.Copy().MarginLeft(1).Width(3).Render(getStringFirstLetter(i.Name))
		}
		userName := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e5e4e2")).Render(i.UserName)
		name := lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e4e2")).Render(i.Name)
		channelMemberItem := channelNameStyle.Copy().Bold(false).Align(lipgloss.Left).UnsetPaddingTop().UnsetMarginTop().MaxWidth(120).Render(activeStatus + "   " + nameLetterChat + " " + userName + " " + name)
		if index == m.Index() {
			userName := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#119da4")).Render(i.UserName)
			name := lipgloss.NewStyle().Foreground(lipgloss.Color("#119da4")).Render(i.Name)
			channelMemberItem = channelNameStyle.Copy().Align(lipgloss.Left).UnsetPaddingTop().UnsetMarginTop().Foreground(lipgloss.Color("#119da4")).Render(activeStatus + "   " + nameLetterChat + " " + userName + " " + name)
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
	m.channelList.Styles.Title = titleStyle.Width(m.width/4 - 1)
	m.channelList.Styles.PaginationStyle = paginationStyle
	m.channelList.Styles.HelpStyle = helpStyle

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
	messagePagesCount := channelOptionsButtonStyle.Width((3 * m.width / 8)).Render(nameLetterBoxStyle.Background(lipgloss.Color("#13505b")).Render(strconv.Itoa(m.messagesList.Paginator.Page+1) + "/" + strconv.Itoa((m.messagesList.Paginator.TotalPages))))
	channelWindowTopbar := channelWindowTopbarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, channelWindowTitle, messagePagesCount))

	channelConversationScreen := lipgloss.NewStyle().Height(m.height - cmndboxSpace).MaxHeight(m.height - cmndboxSpace).Render(m.messagesList.View())

	var smallListBox string
	if m.showSlashCommandList {
		smallListBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, false, true).BorderForeground(lipgloss.Color("#119da4")).Align(lipgloss.Left).Width((3 * m.width / 4) - 4).Render(m.slashCommandsList.View())
	} else if m.showChannelMembersList {
		smallListBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, false, true).BorderForeground(lipgloss.Color("#119da4")).Align(lipgloss.Left).Width((3 * m.width / 4) - 4).Render(m.channelMembersList.View())
	}

	messageEmojiIcon := messageEmojiIconStyle.Render("☺")

	var channelMessageInputBox string
	if m.typing {
		channelMessageInputBox = channelMessageInputBoxStyle.Copy().BorderForeground(lipgloss.Color("#ffffff")).Width((3 * m.width / 4) - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, messageEmojiIcon, m.textInput.View()))
	} else {
		channelMessageInputBox = channelMessageInputBoxStyle.Width((3 * m.width / 4) - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, messageEmojiIcon, m.textInput.View()))
	}

	channelWindow := lipgloss.Place(3*(m.width/4)-2, m.height-2,
		lipgloss.Left, lipgloss.Top,
		channelWindowStyle.Height(m.height-2).Width(3*(m.width/4)).Render(lipgloss.JoinVertical(lipgloss.Top, channelWindowTopbar, channelConversationScreen, smallListBox, channelMessageInputBox)),
	)

	instruction := instructionStyle.Width(m.width).Render("Press Ctrl + C - quit • Ctrl + H - help • Ctrl + Arrows - for navigation in pane • Enter - send message • Ctrl + L - Log out")
	ui := lipgloss.JoinHorizontal(lipgloss.Center, sidebar, channelWindow)
	completeUi := lipgloss.JoinVertical(lipgloss.Center, instruction, ui)
	return completeUi
}

func (m *Model) RenderLoginScreen() string {
	welcomeText := lipgloss.NewStyle().Width(46).Align(lipgloss.Center).Foreground(lipgloss.Color("#767373")).MarginTop(1).Render("WELCOME TO ROCKET.CHAT")
	loginHeadingText := lipgloss.NewStyle().Width(46).Align(lipgloss.Center).Foreground(lipgloss.Color("#cbcbcb")).MarginTop(1).Bold(true).Underline(true).Render("Login into your Account")

	emailInputHeading := lipgloss.NewStyle().Width(46).Align(lipgloss.Left).MarginTop(2).Foreground(lipgloss.Color("#cbcbcb")).Bold(true).Render("E-mail")
	var emailInputBox string
	if m.loginScreen.activeElement == 1 {
		emailInputBox = lipgloss.NewStyle().Width(46).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#cbcbcb")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left).Render(m.loginScreen.emailInput.View())
	} else {
		emailInputBox = lipgloss.NewStyle().Width(46).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#767373")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left).Render(m.loginScreen.emailInput.View())
	}

	passowrdInputHeading := lipgloss.NewStyle().Width(46).Align(lipgloss.Left).MarginTop(1).Foreground(lipgloss.Color("#cbcbcb")).Bold(true).Render("Password")
	var passowrdInputBox string
	if m.loginScreen.activeElement == 2 {
		passowrdInputBox = lipgloss.NewStyle().Width(46).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#cbcbcb")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left).Render(m.loginScreen.passwordInput.View())
	} else {
		passowrdInputBox = lipgloss.NewStyle().Width(46).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#767373")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left).Render(m.loginScreen.passwordInput.View())
	}

	var loginButton string
	if m.loginScreen.activeElement == 3 {
		loginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#00686D")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginBottom(1).MarginTop(2).Render("LOG INTO TUI")
	} else {
		loginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginBottom(1).MarginTop(2).Render("LOG INTO TUI")
	}

	loginUiBox := lipgloss.NewStyle().Align(lipgloss.Center).BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("#119da4")).Height(15).Width(50).Render(lipgloss.JoinVertical(lipgloss.Top, welcomeText, loginHeadingText, emailInputHeading, emailInputBox, passowrdInputHeading, passowrdInputBox, loginButton))

	return loginUiBox
}
