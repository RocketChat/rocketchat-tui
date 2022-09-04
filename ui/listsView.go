package ui

import (
	"fmt"
	"io"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/rocketchat-tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// All lists will be of this type.
type ListDelegate interface {
	Height() int
	Spacing() int
	Update(msg tea.Msg, m *list.Model) tea.Cmd
	Render(w io.Writer, m list.Model, index int, messageListItem list.Item)
}

// Messages are shown as a list.
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

	fmt.Fprint(w, userMessageBox)
}

// Channels are shown using this list.
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

		fmt.Fprint(w, str)
	}
}

// Slash commands are shown using this list.
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
		fmt.Fprint(w, slashCommand)
	}

}

// Channel members are shown using this list while mentioning them in message.
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
		fmt.Fprint(w, channelMemberItem)
	}

}