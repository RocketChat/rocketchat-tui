package ui

import (
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/rest"
	"github.com/RocketChat/rocketchat-tui/cache"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) connectFromEmailAndPassword() error {

	sUrl := m.serverUrl
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
	cache.CreateUpdateCacheEntry("token", user.Token)
	cache.CreateUpdateCacheEntry("tokenExpires", strconv.Itoa(int(user.TokenExpires)))

	m.loginScreen.loggedIn = true
	m.loginScreen.loginScreenState = "showTui"

	m.getSubscriptions()
	return nil
}

func (m *Model) connectFromToken() error {

	sUrl := m.serverUrl
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

	m.getSubscriptions()
	return nil
}

func (m *Model) userLoginBegin() tea.Cmd {
	token, err := cache.GetCacheEntry("token")
	if err != nil {
		log.Println(err)
		m.loginScreen.loginScreenState = "showLoginScreen"
		m.loginScreen.loggedIn = false
		return nil
	}
	tokenExpiresTime, err := cache.GetCacheEntry("tokenExpires")
	if err != nil {
		log.Println(err)
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
		cache.CreateUpdateCacheEntry("token", "")
		cache.CreateUpdateCacheEntry("tokenExpires", "")
		m.loginScreen.loginScreenState = "showLoginScreen"
		m.loginScreen.loggedIn = false
		return nil
	}

}

func (m *Model) handleLoginScreenUpdate(msg tea.Msg) tea.Cmd {
	if !m.loginScreen.loggedIn && m.loginScreen.loginScreenState == "showLoginScreen" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return tea.Quit
			case "tab", "ctrl+down":
				if m.loginScreen.activeElement < 3 {
					m.loginScreen.activeElement = m.loginScreen.activeElement + 1
				} else {
					m.loginScreen.activeElement = 1
				}
				return nil
			case "enter":
				m.loginScreen.activeElement = 3
				if m.email != "" && m.password != "" {
					err := m.connectFromEmailAndPassword()
					if err != nil {
						log.Println("Error occurred while login from email", err)
						panic(err)
					}
					var cmds []tea.Cmd
					channelCmd := m.setChannelsInUiList()
					cmd := m.waitForIncomingMessage(m.msgChannel)
					m.loginScreen.loggedIn = true
					m.changeSelectedChannel(0)
					m.typing = true
					setSlashCommandsList := m.fetchAllSlashCommands()
					channelMembersSetCmnd := m.getChannelMembers()
					cmds = append(cmds, channelCmd, textinput.Blink, setSlashCommandsList, channelMembersSetCmnd, cmd)
					return tea.Batch(cmds...)
				}
			}
		}

		if m.loginScreen.activeElement == 1 {
			var cmd tea.Cmd
			m.loginScreen.emailInput, cmd = m.loginScreen.emailInput.Update(msg)
			m.email = m.loginScreen.emailInput.Value()
			return cmd
		}

		if m.loginScreen.activeElement == 2 {
			var cmd tea.Cmd
			m.loginScreen.passwordInput, cmd = m.loginScreen.passwordInput.Update(msg)
			m.password = m.loginScreen.passwordInput.Value()
			return cmd
		}
	}
	return nil
}

func (m *Model) handleUserLogOut() (tea.Model, tea.Cmd) {
	sUrl := m.serverUrl
	m = IntialModelState(sUrl)
	m.loginScreen.passwordInput.Reset()
	m.loginScreen.emailInput.Reset()
	m.email = ""
	m.password = ""
	m.token = ""
	m.loginScreen.loginScreenState = "showLoginScreen"
	m.loginScreen.loggedIn = false
	cache.CreateUpdateCacheEntry("token", "")
	cache.CreateUpdateCacheEntry("tokenExpires", "")
	return m, nil
}
