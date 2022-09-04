package ui

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	tea "github.com/charmbracelet/bubbletea"
)

// It is used to fetch available slash commands to list them in the UI.
// It supports the offset, count and Sort Query Parameters.
func (m *Model) fetchAllSlashCommands() ([]models.SlashCommand, error) {
	params := url.Values{}
	params.Add("offset", "0")
	resp, err := m.restClient.GetSlashCommandsList(params)
	if err != nil {
		log.Println(err, resp)
		return nil, err
	}
	return resp, nil
}

// It is used to set available slash commands in the model state.
// It is also used to set those commands in the slash commands list so that it can be rendered in the UI.
func (m *Model) setSlashCommandsList() tea.Cmd {
	resp, err := m.fetchAllSlashCommands()
	if err != nil {
		log.Println(err)
		return nil
	}
	var items []list.Item
	m.slashCommands = resp
	for _, cmd := range resp {
		if cmd.Command != "" {
			items = append(items, SlashCommandsItem(cmd))
		}
	}
	slashCommandsList := m.slashCommandsList.SetItems(items)
	return slashCommandsList
}

// It is used to execute a slash command.
// It room id of the subscribed channel, command and params are required to execute it.
func (m *Model) executeSlashCommand(cmnd string, params string) error {
	resp, err := m.restClient.ExecuteSlashCommand(&m.activeChannel, cmnd, params)
	if err != nil {
		log.Println(err, resp)
		return err
	}
	return nil
}

// It is used to filter list of  slash command while typing them.
// Bubble tea provides Filter function which will give us ranks of matching items through fuzzy search by default and their Index in the original list.
// Index obtained from the ranks are used to make list of slash commands as per ranking and then showing  that list in the TUI.
// If user typed command doesn't match any in the slash commands list ranks array will have zero elements and hence we will update the list with the original slash commands.
func (m *Model) handleFilteringSlashCommandList(val string) (tea.Model, tea.Cmd) {
	var slashCmnds []string
	for _, c := range m.slashCommands {
		slashCmnds = append(slashCmnds, c.Command)
	}
	splittedString := strings.Split(val, "/")
	typedCmnd := strings.Split(splittedString[1], " ")
	ranks := m.slashCommandsList.Filter(typedCmnd[0], slashCmnds)
	if typedCmnd[0] != m.selectedSlashCommand.Command && len(ranks) != 0 {
		m.showSlashCommandList = true
		if len(ranks) > 0 {
			var items []list.Item
			for _, rank := range ranks {
				items = append(items, SlashCommandsItem(m.slashCommands[rank.Index]))
			}
			slashCommandsList := m.slashCommandsList.SetItems(items)
			return m, slashCommandsList
		}
	}
	if len(ranks) == 0 {
		m.showSlashCommandList = false
		m.selectedSlashCommand = &models.SlashCommand{}

		var items []list.Item
		for _, cmd := range m.slashCommands {
			if cmd.Command != "" {
				items = append(items, SlashCommandsItem(cmd))
			}
		}
		slashCommandsList := m.slashCommandsList.SetItems(items)
		return m, slashCommandsList
	} else {
		return m, nil
	}
}

// It will handle when to show slash command list.
// Slash command list should appear when first textinput character is '/' and character just after it is not space.
// If the typed slash command is present in filtered slash command list, list will be shown. If not we will hide not display list.
// In all other casese when user is typing message we will not show slash command list.
func (m *Model) handleShowingSlashCommandList() (tea.Model, tea.Cmd) {
	val := m.textInput.Value()
	chars := []rune(val)

	if val == "/" {
		m.showSlashCommandList = true
		return m, nil
	}
	if len(chars) > 1 && string(chars[0]) == "/" && string(chars[1]) != " " {
		m, slashCommandsListCmd := m.handleFilteringSlashCommandList(val)
		return m, slashCommandsListCmd
	} else {
		m.showSlashCommandList = false
		m.selectedSlashCommand = &models.SlashCommand{}

		var items []list.Item
		for _, cmd := range m.slashCommands {
			if cmd.Command != "" {
				items = append(items, SlashCommandsItem(cmd))
			}
		}
		slashCommandsList := m.slashCommandsList.SetItems(items)
		return m, slashCommandsList
	}
}

// It will select slash command and set selected slash command in model state.
// When user again press enter and has not change selected slash command that slash command will be executed in all other cases normal message will be sent.
func (m *Model) handleMessageAndSlashCommandInput() (tea.Model, tea.Cmd) {
	value := m.textInput.Value()
	chars := []rune(value)

	if m.showSlashCommandList {
		cmnd, ok := m.slashCommandsList.SelectedItem().(SlashCommandsItem)
		if ok {
			m.textInput.SetValue("/" + cmnd.Command + " ")
			m.textInput.CursorEnd()
			m.selectedSlashCommand = &models.SlashCommand{
				Command:         cmnd.Command,
				Params:          cmnd.Params,
				Description:     cmnd.Description,
				ProvidesPreview: cmnd.ProvidesPreview,
				ClientOnly:      cmnd.ClientOnly,
			}
		}
		m.showSlashCommandList = false
		return m, nil
	}
	if len(chars) > 1 && string(chars[0]) == "/" && string(chars[1]) != " " && m.selectedSlashCommand.Command != "" {
		splittedString := strings.Split(value, "/")
		typedCmnd := strings.Split(splittedString[1], " ")
		if typedCmnd[0] == m.selectedSlashCommand.Command {
			params := strings.Join(typedCmnd[1:], " ")
			err := m.executeSlashCommand(m.selectedSlashCommand.Command, params)
			if err != nil {
				panic(err)
			}
			m.textInput.Reset()
			m.selectedSlashCommand = &models.SlashCommand{}
			return m, nil
		}
	}
	m.handleMessageSending()
	return m, nil
}

// It is used fetch list of members in a channel or group
func (m *Model) fetchChannelMembers() ([]models.User, error) {
	channel := &models.Channel{
		ID:   m.activeChannel.RoomId,
		Name: m.activeChannel.Name,
	}

	resp, err := m.restClient.GetGroupMembers(channel)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// It is used tob set channel members in the list which will be used for @ mentioning.
// As we use 'all' and 'here' too in mentioning it is also added in the list.
func (m *Model) setChannelMembersList() tea.Cmd {
	resp, err := m.fetchChannelMembers()
	if err != nil {
		log.Println(err, resp)
		return nil
	}
	var items []list.Item
	mentionAll := models.User{
		UserName:     "all",
		Name:         "Notify all in this room",
		Status:       "",
		TokenExpires: 0,
	}
	mentionOnline := models.User{
		UserName:     "here",
		Name:         "Notify active users in this room",
		Status:       "",
		TokenExpires: 0,
	}
	items = append(items, ChannelMembersItem(mentionAll), ChannelMembersItem(mentionOnline))
	m.channelMembers = append(m.channelMembers, mentionAll, mentionOnline)

	for _, u := range resp {
		if u.UserName != "" {
			items = append(items, ChannelMembersItem(u))
			m.channelMembers = append(m.channelMembers, u)
		}
	}
	channelMembersListSetCmnd := m.channelMembersList.SetItems(items)
	return channelMembersListSetCmnd
}

func (m *Model) handleShowingChannelMembersList() tea.Cmd {
	val := m.textInput.Value()
	chars := []rune(val)
	cursorCurrentPos := m.textInput.Cursor()

	if m.showChannelMembersList &&
		m.positionOfAtSymbol <= len(val) &&
		m.positionOfAtSymbol != -1 &&
		cursorCurrentPos >= 1 &&
		string(chars[cursorCurrentPos-1]) != " " { // To filter list only when showing of list is true and position of '@' symbol is less than equal to length of input string (edge case is when '@' is at first position - index will be out of range) and position of '@' symbol is not -1 and cursor current position is greater than equal one and user has not pressed space while entering name for filtering
		username := stringUsernameExtractor(val, m.positionOfAtSymbol)
		cmd := m.handleChannelMemberListFiltering(username)
		return cmd
	} else if cursorCurrentPos == 1 && string(chars[0]) == "@" { // When '@' is at first position of input string
		m.showChannelMembersList = true
		m.positionOfAtSymbol = 1
		return nil
	} else if cursorCurrentPos >= 2 && string(chars[cursorCurrentPos-1]) == "@" && string(chars[cursorCurrentPos-2]) == " " { // When '@' is at position other than first just after space
		m.showChannelMembersList = true
		m.positionOfAtSymbol = cursorCurrentPos
		return nil
	} else { // In all other case don't show list, add original members in the list
		m.positionOfAtSymbol = -1
		var items []list.Item
		for _, u := range m.channelMembers {
			items = append(items, ChannelMembersItem(u))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		m.showChannelMembersList = false
		return filteredChannelMembersList
	}
}

func (m *Model) handleChannelMemberListFiltering(username string) tea.Cmd {
	if len(username) == 0 {
		var items []list.Item
		for _, u := range m.channelMembers {
			items = append(items, ChannelMembersItem(u))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		return filteredChannelMembersList
	}
	var usernames []string
	for _, u := range m.channelMembers {
		usernames = append(usernames, u.UserName)
	}
	ranks := m.channelMembersList.Filter(username, usernames)
	if len(ranks) > 0 {
		var items []list.Item
		for _, rank := range ranks {
			items = append(items, ChannelMembersItem(m.channelMembers[rank.Index]))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		return filteredChannelMembersList
	} else {
		var items []list.Item
		for _, u := range m.channelMembers {
			items = append(items, ChannelMembersItem(u))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		m.showChannelMembersList = false
		return filteredChannelMembersList
	}
}

func (m *Model) handleSelectingAtChannelMember() tea.Cmd {
	diff := m.textInput.Cursor() - m.positionOfAtSymbol
	if m.showChannelMembersList {
		user, ok := m.channelMembersList.SelectedItem().(ChannelMembersItem)
		if ok {
			userMessageString := usernameAutoCompleteString(m.textInput.Value(), user.UserName+" ", m.positionOfAtSymbol, diff)
			m.textInput.SetValue(userMessageString)
			m.textInput.SetCursor(m.positionOfAtSymbol + len(user.UserName))
		}
		m.showChannelMembersList = false
		m.positionOfAtSymbol = -1

		var items []list.Item
		for _, u := range m.channelMembers {
			items = append(items, ChannelMembersItem(u))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		return filteredChannelMembersList
	}
	return nil
}

// It contains all the key press events and the update events to perform after them.
// It will update the UI as according to the key pressed.
func (m *Model) handleUpdateOnKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.ChannelListNextChannel):
		m.channelList.CursorDown()
		return m, nil
	case key.Matches(msg, m.keys.ChannelListPreviousChannel):
		m.channelList.CursorUp()
		return m, nil
	case key.Matches(msg, m.keys.MessageListNextPage):
		m.messagesList.Paginator.NextPage()
		return m, nil
	case key.Matches(msg, m.keys.MessageListPreviousPage):
		m.messagesList.Paginator.PrevPage()
		if m.messagesList.Paginator.Page == 0 && m.loadMorePastMessages {
			m.loadMorePastMessages = false
			msgsCmd := m.fetchPastMessages()
			return m, msgsCmd
		}
		if m.messagesList.Paginator.Page == 0 {
			m.loadMorePastMessages = true
		}
		return m, nil
	case key.Matches(msg, m.keys.SlashCommandListNextCommand) && m.showSlashCommandList:
		m.slashCommandsList.CursorDown()
		return m, nil
	case key.Matches(msg, m.keys.SlashCommandListPreviousCommand) && m.showSlashCommandList:
		m.slashCommandsList.CursorUp()
		return m, nil
	case key.Matches(msg, m.keys.ChannelMembersListNextMember) && m.showChannelMembersList:
		m.channelMembersList.CursorDown()
		return m, nil
	case key.Matches(msg, m.keys.ChannelMembersListPreviousMember) && m.showChannelMembersList:
		m.channelMembersList.CursorUp()
		return m, nil
	case key.Matches(msg, m.keys.QuitAndCloseTui):
		return m, tea.Quit
	case key.Matches(msg, m.keys.SelectByEnterKeyPress):
		if !m.typing {
			m.typing = true
			var msgItems []list.Item
			cmd := m.messagesList.SetItems(msgItems)
			m.changeSelectedChannel(m.channelList.Index())
			m.loadMorePastMessages = false
			return m, cmd
		}

		if m.typing {
			if m.showChannelMembersList {
				channelMemberCmd := m.handleSelectingAtChannelMember()
				return m, channelMemberCmd
			}
			m, cmd := m.handleMessageAndSlashCommandInput()
			return m, cmd
		}
	case key.Matches(msg, m.keys.MessageTypingInactive):
		m.typing = !m.typing
		return m, nil
	case key.Matches(msg, m.keys.LogOutTui):
		m, cmd := m.handleUserLogOut()
		return m, cmd
	}
	if m.typing {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		channelMembersCmnd := m.handleShowingChannelMembersList()
		m, slashCmnd := m.handleShowingSlashCommandList()
		return m, tea.Batch(cmd, slashCmnd, channelMembersCmnd)
	}
	return m, nil
}

// To search any user while mentioning them in room.
func (m *Model) handleSearchUser(query string) ([]models.SearchUsers, error) {
	resp, err := m.restClient.SearchUsersOrRooms(query)
	b, _ := json.Marshal(resp)
	log.Println(string(b))
	if err != nil {
		log.Println(err, resp)
		return nil, err
	}
	users := resp.Users
	return users, nil
}
