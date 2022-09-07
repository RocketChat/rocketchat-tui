package keyBindings

import "github.com/charmbracelet/bubbles/key"

// All the key bindings required for performing different actions in the TUI.
type ListKeyMap struct {
	MessageListNextPage              key.Binding
	MessageListPreviousPage          key.Binding
	ChannelListNextChannel           key.Binding
	ChannelListPreviousChannel       key.Binding
	SlashCommandListNextCommand      key.Binding
	SlashCommandListPreviousCommand  key.Binding
	ChannelMembersListNextMember     key.Binding
	ChannelMembersListPreviousMember key.Binding
	QuitAndCloseTui                  key.Binding
	SelectByEnterKeyPress            key.Binding
	MessageTypingInactive            key.Binding
	LogOutTui                        key.Binding
}

// For binding keyboard keys with the appropriate key bindings.
func NewListKeyMap() *ListKeyMap {
	return &ListKeyMap{
		MessageListNextPage: key.NewBinding(
			key.WithKeys("ctrl+right"),
			key.WithHelp("ctrl+right", "Next Messages Page"),
		),
		MessageListPreviousPage: key.NewBinding(
			key.WithKeys("ctrl+left"),
			key.WithHelp("ctrl+left", "Previous Messages Page"),
		),
		ChannelListNextChannel: key.NewBinding(
			key.WithKeys("ctrl+down"),
			key.WithHelp("ctrl+down", "Next Channel"),
		),
		ChannelListPreviousChannel: key.NewBinding(
			key.WithKeys("ctrl+up"),
			key.WithHelp("ctrl+up", "Previous Channel"),
		),
		SlashCommandListNextCommand: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "Next Slash Command"),
		),
		SlashCommandListPreviousCommand: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "Previous Slash Command"),
		),
		ChannelMembersListNextMember: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "Next Channel Member"),
		),
		ChannelMembersListPreviousMember: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "Previous Channel Member"),
		),
		QuitAndCloseTui: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "Quit from Tui"),
		),
		SelectByEnterKeyPress: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "To select channel and send message"),
		),
		MessageTypingInactive: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Inactive message typing input"),
		),
		LogOutTui: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "Log out profile from Tui"),
		),
	}
}