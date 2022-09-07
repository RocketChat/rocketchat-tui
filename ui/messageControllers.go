package ui

import (
	"log"
	"strings"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// It calls the Realtime API function used to send message in the TUI.
func (m *Model) sendMessage(text string) {
	if text != "" {
		channelId := m.activeChannel.RoomId
		if _, err := m.rlClient.SendMessage(&models.Message{RoomID: channelId, Msg: text}); err != nil {
			log.Println(err)
		}
	}
}

// It calls the Realtime API function to load past message history of a room when the TUI first rendered.
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

// It calls the REST API function to fetch more past messages of a romm.
// It is called when user want to load more past message.
// It calls the appropriate API according to the type of channel from public (channel), private (group) and DM
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
			log.Println("CHANNEL MESSAGE ERROR", err)
		}
	case "d":
		messages, err = m.restClient.DMHistory(channel, true, *today, page)
		if err != nil {
			log.Println("DIRECT MESSAGE ERROR", err)
		}
	default:
		messages, err = m.restClient.GroupHistory(channel, true, *today, page)
		if err != nil {
			log.Println("GROUP MESSAGE ERROR", err)
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
		msgsListItems = append(msgsListItems, MessagessItem(msg))
	}
	msgsCommand := m.messagesList.SetItems(msgsListItems)
	return msgsCommand
}

// It is used for the realtime updation of chat messages in the TUI.
// It return a 'tea.Cmd' which is a function which returns 'tea.Msg' as it triggers the Update function.
// The 'tea.Msg' here returned will be of type models.Message which is catched in TUI Update function and hence TUI is updated with new message.
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

// It is used for handling sending of message from the TUI and check message before sending.
func (m *Model) handleMessageSending() {
	msg := strings.TrimSpace(m.textInput.Value())
	if msg != "" {
		m.sendMessage(msg)
		m.textInput.Reset()
		return
	}
	m.textInput.Reset()
}
