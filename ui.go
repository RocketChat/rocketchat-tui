package main

import (
	"fmt"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
)

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

	// var userMessageBox string
	// if index == m.Index() {
	// 	userMessageBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderBackground(lipgloss.Color("#ffffff")).PaddingBottom(1).Width(80).Render(lipgloss.JoinHorizontal(lipgloss.Left, nameLetterChat, messageBox))
	// } else {
	// 	userMessageBox = lipgloss.NewStyle().PaddingBottom(1).Width(80).Render(lipgloss.JoinHorizontal(lipgloss.Left, nameLetterChat, messageBox))

	// }
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

func (m *Model) RenderTui() string {
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

	var channelMessageInputBox string
	if m.typing {
		channelMessageInputBox = channelMessageInputBoxStyle.Copy().BorderForeground(lipgloss.Color("#ffffff")).Width((3 * m.width / 4) - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, messageEmojiIcon, m.textInput.View()))
	} else {
		channelMessageInputBox = channelMessageInputBoxStyle.Width((3 * m.width / 4) - 4).Render(lipgloss.JoinHorizontal(lipgloss.Top, messageEmojiIcon, m.textInput.View()))
	}

	channelWindow := lipgloss.Place(3*(m.width/4)-2, m.height-2,
		lipgloss.Left, lipgloss.Top,
		channelWindowStyle.Height(m.height-2).Width(3*(m.width/4)).Render(lipgloss.JoinVertical(lipgloss.Top, channelWindowTopbar, channelConversationScreen, channelMessageInputBox)),
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

	// authTokenInputHeading := lipgloss.NewStyle().Align(lipgloss.Left).MarginTop(5).Foreground(lipgloss.Color("#cbcbcb")).Bold(true).Render("Password")
	// authTokenInputBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#cbcbcb")).Foreground(lipgloss.Color("#1b1b1b")).Align(lipgloss.Left).Render(m.loginScreen.authTokenInput.View())

	var loginButton string
	if m.loginScreen.activeElement == 3 {
		loginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#00686D")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginBottom(1).MarginTop(2).Render("LOG INTO TUI")
	} else {
		loginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginBottom(1).MarginTop(2).Render("LOG INTO TUI")
	}

	// orText := lipgloss.NewStyle().Width(46).Align(lipgloss.Center).Foreground(lipgloss.Color("#767373")).MarginTop(1).Render("OR")

	// var authLoginButton string
	// if m.loginScreen.activeElement == 4 {
	// 	authLoginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#00686D")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginTop(1).MarginBottom(1).Render("LOG IN USING AUTH TOKEN")
	// } else {
	// 	authLoginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginTop(1).MarginBottom(1).Render("LOG IN USING AUTH TOKEN")
	// }

	loginUiBox := lipgloss.NewStyle().Align(lipgloss.Center).BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("#119da4")).Height(15).Width(50).Render(lipgloss.JoinVertical(lipgloss.Top, welcomeText, loginHeadingText, emailInputHeading, emailInputBox, passowrdInputHeading, passowrdInputBox, loginButton))

	return loginUiBox
}
