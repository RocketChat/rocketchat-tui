package ui

import (
	"log"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) fetchAllSlashCommands() tea.Cmd {
	params := url.Values{}
	params.Add("offset", "0")
	resp, err := m.restClient.GetSlashCommandsList(params)
	if err != nil {
		log.Println(err, resp)
		panic(err)
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

func (m *Model) executeSlashCommand(cmnd string, params string) error {
	resp, err := m.restClient.ExecuteSlashCommand(&m.activeChannel, cmnd, params)
	if err != nil {
		log.Println(err, resp)
		return err
	}
	return nil
}

func (m *Model) handleShowingAndFilteringSlashCommandList() (tea.Model, tea.Cmd) {
	val := m.textInput.Value()
	chars := []rune(val)

	if val == "/" {
		m.showSlashCommandList = true
		return m, nil
	}
	if len(chars) > 1 && string(chars[0]) == "/" && string(chars[1]) != " " {
		var slashCmnds []string
		for _, c := range m.slashCommands {
			slashCmnds = append(slashCmnds, c.Command)
		}
		splittedString := strings.Split(val, "/")
		typedCmnd := strings.Split(splittedString[1], " ")
		ranks := m.slashCommandsList.Filter(typedCmnd[0], slashCmnds)
		if typedCmnd[0] != m.selectedSlashCommand.Command {
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
		}
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
	return m, nil
}

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
	if len(chars) > 1 && string(chars[0]) == "/" && string(chars[1]) != " " {
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
		} else {
			m, cmd := m.handleMessageSending()
			return m, cmd
		}
	} else {
		m, cmd := m.handleMessageSending()
		return m, cmd
	}
}

func (m *Model) getChannelMembers() tea.Cmd {
	channel := &models.Channel{
		ID:   m.activeChannel.RoomId,
		Name: m.activeChannel.Name,
	}

	resp, err := m.restClient.GetGroupMembers(channel)
	if err != nil {
		log.Println(err, resp)
		panic(err)
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
	items = append(items, ChannelMembersItem(mentionAll))
	items = append(items, ChannelMembersItem(mentionOnline))
	m.channelMembers = append(m.channelMembers, mentionAll)
	m.channelMembers = append(m.channelMembers, mentionOnline)

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

	if m.showChannelMembersList && m.positionOfAtSymbol <= len(val) && m.positionOfAtSymbol != -1 && string(chars[m.positionOfAtSymbol-1]) == "@" {
		username := stringUsernameExtractor(val, m.positionOfAtSymbol)
		cmd := m.handleChannelMemberListFiltering(username)
		return cmd

	} else if m.positionOfAtSymbol == cursorCurrentPos {
		m.showChannelMembersList = true
		return nil
	} else if cursorCurrentPos == 1 && string(chars[0]) == "@" {
		m.showChannelMembersList = true
		m.positionOfAtSymbol = 1
		return nil
	} else if cursorCurrentPos >= 2 && string(chars[cursorCurrentPos-1]) == "@" && string(chars[cursorCurrentPos-2]) == " " {
		m.showChannelMembersList = true
		m.positionOfAtSymbol = cursorCurrentPos
		return nil
	} else {
		m.positionOfAtSymbol = -1
		m.showChannelMembersList = false
		return nil
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
			m.textInput.CursorEnd()
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

// func userNameInMsgHighlighter(msg string) string {
// 	if len(msg) == 0 {
// 		return msg
// 	}
// 	str := strings.Split(msg, " ")
// 	for _, v := range str {
// 		word := []rune(v)
// 		if len(word) > 0 && string(word[0]) == "@" {
// 			// isUserInChannel := doesUserExistInChannel()
// 		}
// 	}
// 	return ""
// }

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
		m, slashCmnd := m.handleShowingAndFilteringSlashCommandList()
		return m, tea.Batch(cmd, slashCmnd, channelMembersCmnd)
	}
	return m, nil
}
