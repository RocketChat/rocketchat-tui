package main

import (
	// "encoding/json"
	"log"
	"net/url"
	"os"
	"time"

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
	CreateUpdateCacheEntry("tokenGeneratedTime", time.Now().Format(time.RFC3339))

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
	tokenGeneratedTime, _ := GetCacheEntry("tokenGeneratedTime")

	tokenValid := CheckForTokenExpiration(tokenGeneratedTime)
	if tokenValid {
		m.token = token
		err := m.connectFromToken()
		if err != nil {
			os.Exit(1)
		}
		// go m.handleMessageStream()

		channelCmd := m.setChannelsInUiList()
		m.changeSelectedChannel(0)
		return channelCmd

	} else {
		CreateUpdateCacheEntry("token", "")
		CreateUpdateCacheEntry("tokenGeneratedTime", "")
		m.loginScreen.loginScreenState = "showLoginScreen"
		m.loginScreen.loggedIn = false
		return nil
	}

}

func (m *Model) changeSelectedChannel(index int) {
	m.activeChannel = m.subscriptionList[index]

	m.messageHistory = []models.Message{}
	messageList = []models.Message{}

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

}

func (m *Model) getSubscriptions() {

	// Sonyflake, DDP, gabs
	subscriptions, err := m.rlClient.GetChannelSubscriptions()
	if err != nil {
		panic(err)
	}

	for _, sub := range subscriptions {
		if sub.Open && sub.Name != "" {
			m.subscriptionList = append(m.subscriptionList, sub)
		}
	}

	m.loadChannels = true

	// bs, _ := json.Marshal(m.subscriptionList)
	// log.Println(string(bs))
}

func (m *Model) setChannelsInUiList() tea.Cmd {
	var items []list.Item
	for _, sub := range m.subscriptionList {
		if sub.Open && sub.Name != "" {
			items = append(items, channelsItem(sub))
		}
	}
	// PrintToLogFile(m.messageHistory)
	channelCmd := m.channelList.SetItems(items)
	m.loadChannels = false
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
	CreateUpdateCacheEntry("tokenGeneratedTime", "")
	return m, nil
}
