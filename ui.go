package main

import "github.com/charmbracelet/lipgloss"

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

	instruction := instructionStyle.Width(m.width).Render("Press Ctrl + C - quit • Ctrl + H - help • Arrows - for navigation in pane • Enter - send message")
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
		loginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#00686D")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginTop(2).Render("LOG INTO TUI")
	} else {
		loginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginTop(2).Render("LOG INTO TUI")
	}

	orText := lipgloss.NewStyle().Width(46).Align(lipgloss.Center).Foreground(lipgloss.Color("#767373")).MarginTop(1).Render("OR")

	var authLoginButton string
	if m.loginScreen.activeElement == 4 {
		authLoginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#00686D")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginTop(1).MarginBottom(1).Render("LOG IN USING AUTH TOKEN")
	} else {
		authLoginButton = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginTop(1).MarginBottom(1).Render("LOG IN USING AUTH TOKEN")
	}

	loginUiBox := lipgloss.NewStyle().Align(lipgloss.Center).BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("#119da4")).Height(15).Width(50).Render(lipgloss.JoinVertical(lipgloss.Top, welcomeText, loginHeadingText, emailInputHeading, emailInputBox, passowrdInputHeading, passowrdInputBox, loginButton, orText, authLoginButton))

	return loginUiBox
}
