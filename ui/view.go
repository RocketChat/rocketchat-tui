package ui

import (
	"strconv"

	"github.com/RocketChat/rocketchat-tui/styles"
	"github.com/charmbracelet/lipgloss"
)

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
	m.channelMembersList.SetShowTitle(false)
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
