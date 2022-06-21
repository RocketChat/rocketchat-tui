package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	nameLetterBoxStyle = lipgloss.NewStyle().Align(lipgloss.Center).Height(1).Width(1).PaddingLeft(1).PaddingRight(1)

	sidebarTopColumnStyle = lipgloss.NewStyle()
	sidebarTopbarStyle    = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("#119da4"))
	sidebarStyle          = lipgloss.NewStyle().Align(lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderRight(true).BorderForeground(lipgloss.Color("#119da4"))

	channelNameStyle            = lipgloss.NewStyle().Bold(true).PaddingLeft(1)
	starIconStyle               = lipgloss.NewStyle().Bold(true).PaddingLeft(1)
	channelWindowTitleStyle     = lipgloss.NewStyle().Align(lipgloss.Left)
	channelOptionsButtonStyle   = lipgloss.NewStyle().Align(lipgloss.Right)
	channelWindowTopbarStyle    = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("#119da4"))
	channelWindowStyle          = lipgloss.NewStyle().Align(lipgloss.Left)
	instructionStyle            = lipgloss.NewStyle().Height(1).Foreground(lipgloss.Color("#767373")).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("#119da4")).Bold(true).Align(lipgloss.Center)
	dialogStyle                 = lipgloss.NewStyle()
	channelMessageInputBoxStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#119da4")).Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left)
	messageEmojiIconStyle       = lipgloss.NewStyle().PaddingRight(2)
	rocketChatIconStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5757"))

	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("#0c7489")).Foreground(lipgloss.Color("#ffffff")).PaddingLeft(1).Align(lipgloss.Left).Bold(true)
	paginationStyle   = list.DefaultStyles().PaginationStyle
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingBottom(1)
)
