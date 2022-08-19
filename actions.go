package main

import (
	// "encoding/json"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/rest"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) connectFromEmailAndPassword() error {

	sUrl := getServerUrl()
	serverUrl, err := url.Parse(sUrl)
	if err != nil {
		return err
	}

	c, err := realtime.NewClient(serverUrl, false)
	if err != nil {
		log.Println("Failed to connect", err)
		return err
	}

	m.rlClient = c

	user, err := m.rlClient.Login(&models.UserCredentials{Email: m.email, Password: m.password})
	if err != nil {
		return err
	}

	c2 := rest.NewClient(serverUrl, false)

	m.restClient = c2
	if err := m.restClient.Login(&models.UserCredentials{Email: m.email, Password: m.password}); err != nil {
		log.Println("failed to login")
		return err
	}

	m.token = user.Token
	CreateUpdateCacheEntry("token", user.Token)
	CreateUpdateCacheEntry("tokenExpires", strconv.Itoa(int(user.TokenExpires)))

	m.loginScreen.loggedIn = true
	m.loginScreen.loginScreenState = "showTui"

	PrintToLogFile(user)

	// log.Println("BINGO!\nYou are In....")
	m.getSubscriptions()

	// m.handleMessageStream()
	return nil
}

func (m *Model) connectFromToken() error {

	sUrl := getServerUrl()
	serverUrl, err := url.Parse(sUrl)
	if err != nil {
		return err
	}

	c, err := realtime.NewClient(serverUrl, false)
	if err != nil {
		log.Println("Failed to connect", err)
		return err
	}

	m.rlClient = c

	user, err := m.rlClient.Login(&models.UserCredentials{Token: m.token})
	if err != nil {
		return err
	}

	c2 := rest.NewClient(serverUrl, false)

	m.restClient = c2
	if err := m.restClient.Login(&models.UserCredentials{ID: user.ID, Token: m.token}); err != nil {
		log.Println("failed to login")
		return err
	}
	m.loginScreen.loggedIn = true
	m.loginScreen.loginScreenState = "showTui"

	// log.Println("BINGO!\nYou are In....")
	m.getSubscriptions()

	// m.handleMessageStream()
	return nil
}

func (m *Model) userLoginBegin() tea.Cmd {
	token, err := GetCacheEntry("token")
	if err != nil {
		PrintToLogFile(err)
		m.loginScreen.loginScreenState = "showLoginScreen"
		m.loginScreen.loggedIn = false
		return nil
	}
	tokenExpiresTime, err := GetCacheEntry("tokenExpires")
	if err != nil {
		PrintToLogFile(err)
		m.loginScreen.loginScreenState = "showLoginScreen"
		m.loginScreen.loggedIn = false
		return nil
	}

	tokenValid := CheckForTokenExpiration(tokenExpiresTime)
	if tokenValid {
		m.token = token
		err := m.connectFromToken()
		if err != nil {
			os.Exit(1)
		}
		channelCmd := m.setChannelsInUiList()
		m.typing = true
		m.changeSelectedChannel(0)
		setSlashCommandsList := m.fetchAllSlashCommands()
		channelMembersSetCmnd := m.getChannelMembers()
		return tea.Batch(channelCmd, setSlashCommandsList, channelMembersSetCmnd)

	} else {
		CreateUpdateCacheEntry("token", "")
		CreateUpdateCacheEntry("tokenExpires", "")
		m.loginScreen.loginScreenState = "showLoginScreen"
		m.loginScreen.loggedIn = false
		return nil
	}

}

func (m *Model) changeSelectedChannel(index int) {
	m.activeChannel = m.subscriptionList[index]

	m.messageHistory = []models.Message{}

	if _, ok := m.subscribed[m.activeChannel.RoomId]; !ok {
		if err := m.rlClient.SubscribeToMessageStream(&models.Channel{ID: m.activeChannel.RoomId}, m.msgChannel); err != nil {
			log.Println(err)
		}

		m.subscribed[m.activeChannel.RoomId] = m.activeChannel.RoomId
	}

	m.loadHistory()
}

func (m *Model) sendMessage(text string) {
	if text != "" {
		channelId := m.activeChannel.RoomId

		if _, err := m.rlClient.SendMessage(&models.Message{RoomID: channelId, Msg: text}); err != nil {
			log.Println(err)
		}
	}
}

func (m *Model) loadHistory() {
	channelId := m.activeChannel.RoomId

	messages, err := m.rlClient.LoadHistory(channelId)
	if err != nil {
		log.Println(err)
	}

	// Reverse order so will show up properly
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	for _, message := range messages {
		m.msgChannel <- message
	}

	m.lastMessageTimestamp = messages[0].Timestamp
}

func (m *Model) fetchPastMessages() tea.Cmd {
	page := &models.Pagination{
		Count:  20,
		Offset: 5,
		Total:  10,
	}
	channel := &models.Channel{
		ID:    m.activeChannel.RoomId,
		Name:  m.activeChannel.Name,
		Fname: m.activeChannel.DisplayName,
		Type:  m.activeChannel.Type,
	}

	today := m.lastMessageTimestamp
	var (
		messages []models.Message
		err      error
	)

	switch channel.Type {
	case "c":
		messages, err = m.restClient.ChannelHistory(channel, true, *today, page)
		if err != nil {
			PrintToLogFile("CHANNEL MESSAGE ERROR", err)
		}
	case "d":
		messages, err = m.restClient.DMHistory(channel, true, *today, page)
		if err != nil {
			PrintToLogFile("DIRECT MESSAGE ERROR", err)
		}
	default:
		messages, err = m.restClient.GroupHistory(channel, true, *today, page)
		if err != nil {
			PrintToLogFile("GROUP MESSAGE ERROR", err)
		}
	}

	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	var updatedMessageList []models.Message
	updatedMessageList = append(updatedMessageList, messages...)
	updatedMessageList = append(updatedMessageList, m.messageHistory...)

	if len(messages) > 0 {
		m.messageHistory = updatedMessageList
		m.lastMessageTimestamp = messages[0].Timestamp
	}

	var msgsListItems []list.Item
	for _, msg := range updatedMessageList {
		msgsListItems = append(msgsListItems, messagessItem(msg))
	}
	msgsCommand := m.messagesList.SetItems(msgsListItems)
	return msgsCommand
}

func (m *Model) getSubscriptions() {
	subscriptions, err := m.rlClient.GetChannelSubscriptions()
	if err != nil {
		panic(err)
	}

	for _, sub := range subscriptions {
		if sub.Open && sub.Name != "" {
			m.subscriptionList = append(m.subscriptionList, sub)
		}
	}
}

func (m *Model) setChannelsInUiList() tea.Cmd {
	var items []list.Item
	for _, sub := range m.subscriptionList {
		if sub.Open && sub.Name != "" {
			items = append(items, channelsItem(sub))
		}
	}
	channelCmd := m.channelList.SetItems(items)
	m.activeChannel = m.subscriptionList[0]
	return channelCmd
}

func (m *Model) handleUserLogOut() (tea.Model, tea.Cmd) {
	m = IntialModelState()
	m.loginScreen.passwordInput.Reset()
	m.loginScreen.emailInput.Reset()
	m.email = ""
	m.password = ""
	m.token = ""
	m.loginScreen.loginScreenState = "showLoginScreen"
	m.loginScreen.loggedIn = false
	CreateUpdateCacheEntry("token", "")
	CreateUpdateCacheEntry("tokenExpires", "")
	return m, nil
}

func (m *Model) waitForIncomingMessage(msgChannel chan models.Message) tea.Cmd {
	return func() tea.Msg {
		message := <-msgChannel
		if message.RoomID == m.activeChannel.RoomId {
			m.messageHistory = append(m.messageHistory, message)
			return message
		}
		return nil
	}
}

func (m *Model) fetchAllSlashCommands() tea.Cmd {
	params := url.Values{}
	params.Add("offset", "0")
	resp, err := m.restClient.GetSlashCommandsList(params)
	if err != nil {
		PrintToLogFile(err, resp)
		panic(err)
	}
	var items []list.Item
	m.slashCommands = resp
	for _, cmd := range resp {
		if cmd.Command != "" {
			items = append(items, slashCommandsItem(cmd))
		}
	}
	slashCommandsList := m.slashCommandsList.SetItems(items)
	return slashCommandsList
}

func (m *Model) executeSlashCommand(cmnd string, params string) error {
	resp, err := m.restClient.ExecuteSlashCommand(&m.activeChannel, cmnd, params)
	if err != nil {
		PrintToLogFile(err, resp)
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
					items = append(items, slashCommandsItem(m.slashCommands[rank.Index]))
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
					items = append(items, slashCommandsItem(cmd))
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
				items = append(items, slashCommandsItem(cmd))
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
		cmnd, ok := m.slashCommandsList.SelectedItem().(slashCommandsItem)
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

func (m *Model) handleMessageSending() (tea.Model, tea.Cmd) {
	msg := strings.TrimSpace(m.textInput.Value())
	if msg != "" {
		m.sendMessage(msg)
		m.textInput.Reset()
		return m, nil
	} else {
		m.textInput.Reset()
		return m, nil
	}
}

func (m *Model) getChannelMembers() tea.Cmd {
	channel := &models.Channel{
		ID:   m.activeChannel.RoomId,
		Name: m.activeChannel.Name,
	}

	resp, err := m.restClient.GetGroupMembers(channel)
	if err != nil {
		PrintToLogFile(err, resp)
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
	items = append(items, channelMembersItem(mentionAll))
	items = append(items, channelMembersItem(mentionOnline))
	m.channelMembers = append(m.channelMembers, mentionAll)
	m.channelMembers = append(m.channelMembers, mentionOnline)

	for _, u := range resp {
		if u.UserName != "" {
			items = append(items, channelMembersItem(u))
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
			items = append(items, channelMembersItem(u))
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
			items = append(items, channelMembersItem(m.channelMembers[rank.Index]))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		return filteredChannelMembersList
	} else {
		var items []list.Item
		for _, u := range m.channelMembers {
			items = append(items, channelMembersItem(u))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		return filteredChannelMembersList
	}
}

func (m *Model) handleSelectingAtChannelMember() tea.Cmd {
	diff := m.textInput.Cursor() - m.positionOfAtSymbol
	if m.showChannelMembersList {
		user, ok := m.channelMembersList.SelectedItem().(channelMembersItem)
		if ok {
			userMessageString := usernameAutoCompleteString(m.textInput.Value(), user.UserName+" ", m.positionOfAtSymbol, diff)
			m.textInput.SetValue(userMessageString)
			m.textInput.CursorEnd()
		}
		m.showChannelMembersList = false
		m.positionOfAtSymbol = -1

		var items []list.Item
		for _, u := range m.channelMembers {
			items = append(items, channelMembersItem(u))
		}
		filteredChannelMembersList := m.channelMembersList.SetItems(items)
		return filteredChannelMembersList
	}
	return nil
}
