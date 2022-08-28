package ui

import (
	"log"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

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
			items = append(items, ChannelsItem(sub))
		}
	}
	channelCmd := m.channelList.SetItems(items)
	m.activeChannel = m.subscriptionList[0]
	return channelCmd
}
