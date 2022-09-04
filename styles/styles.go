package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// It contain all the styling of the TUI
var (
	NameLetterBoxStyle                = lipgloss.NewStyle().Align(lipgloss.Center).Height(1).Width(1).PaddingLeft(1).PaddingRight(1)
	SidebarTopColumnStyle             = lipgloss.NewStyle()
	SidebarTopbarStyle                = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("#119da4"))
	SidebarStyle                      = lipgloss.NewStyle().Align(lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderRight(true).BorderForeground(lipgloss.Color("#119da4"))
	ChannelNameStyle                  = lipgloss.NewStyle().Bold(true).PaddingLeft(1)
	StarIconStyle                     = lipgloss.NewStyle().Bold(true).PaddingLeft(1)
	ChannelWindowTitleStyle           = lipgloss.NewStyle().Align(lipgloss.Left)
	ChannelOptionsButtonStyle         = lipgloss.NewStyle().Align(lipgloss.Right)
	ChannelWindowTopbarStyle          = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("#119da4"))
	ChannelWindowStyle                = lipgloss.NewStyle().Align(lipgloss.Left)
	InstructionStyle                  = lipgloss.NewStyle().Height(1).Foreground(lipgloss.Color("#767373")).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("#119da4")).Bold(true).Align(lipgloss.Center)
	DialogStyle                       = lipgloss.NewStyle()
	ChannelMessageInputBoxStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left)
	MessageEmojiIconStyle             = lipgloss.NewStyle().PaddingRight(2)
	RocketChatIconStyle               = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5757"))
	TitleStyle                        = lipgloss.NewStyle().Background(lipgloss.Color("#0c7489")).Foreground(lipgloss.Color("#ffffff")).PaddingLeft(1).Align(lipgloss.Left).Bold(true)
	PaginationStyle                   = list.DefaultStyles().PaginationStyle
	HelpStyle                         = list.DefaultStyles().HelpStyle.PaddingBottom(1)
	SmallListBoxStyle                 = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, false, true).BorderForeground(lipgloss.Color("#119da4")).Align(lipgloss.Left)
	LoginScreenWelcomeTextStyle       = lipgloss.NewStyle().Width(46).Align(lipgloss.Center).Foreground(lipgloss.Color("#767373")).MarginTop(1)
	LoginHeadingTextStyle             = lipgloss.NewStyle().Width(46).Align(lipgloss.Center).Foreground(lipgloss.Color("#cbcbcb")).MarginTop(1).Bold(true).Underline(true)
	EmailInputLabelStyle              = lipgloss.NewStyle().Width(46).Align(lipgloss.Left).MarginTop(2).Foreground(lipgloss.Color("#cbcbcb")).Bold(true)
	PasswordInputLabelStyle           = lipgloss.NewStyle().Width(46).Align(lipgloss.Left).MarginTop(1).Foreground(lipgloss.Color("#cbcbcb")).Bold(true)
	LoginScreenInputBoxActiveStyle    = lipgloss.NewStyle().Width(46).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#cbcbcb")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left)
	LoginScreenInputBoxNotActiveStyle = lipgloss.NewStyle().Width(46).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#767373")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left)
	LoginButtonNotActiveStyle         = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginBottom(1).MarginTop(2)
	LoginButtonActiveStyle            = lipgloss.NewStyle().Width(48).Background(lipgloss.Color("#00686D")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Center).Padding(1, 2).MarginBottom(1).MarginTop(2)
	LoginUiBoxStyle                   = lipgloss.NewStyle().Align(lipgloss.Center).BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("#119da4")).Height(15).Width(50)
	ChannelMembersListUsernameStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e5e4e2"))
	ChannelMembersListNameStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#e5e4e2"))
	SlashCommandStyle                 = ChannelNameStyle.Copy().Align(lipgloss.Left).UnsetPaddingTop().UnsetMarginTop().Foreground(lipgloss.Color("#119da4"))
	UserFullNameStyle                 = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).PaddingRight(1).Bold(true).Align(lipgloss.Left)
	UsernameStyle                     = lipgloss.NewStyle().Foreground(lipgloss.Color("#767373")).PaddingRight(1).Bold(true).Align(lipgloss.Left)
	TimestampStyle                    = lipgloss.NewStyle().Foreground(lipgloss.Color("#767373")).Align(lipgloss.Left)
	UserMessageStyle                  = lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#ffffff")).MaxWidth(80).Width(80)
	MessageBoxStyle                   = lipgloss.NewStyle().PaddingLeft(1)
	UserMessageBoxStyle               = lipgloss.NewStyle().PaddingBottom(1).Width(80)
)
