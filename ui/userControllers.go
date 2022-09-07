package ui

import (
	"log"
	"net/url"
	"strconv"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/rest"
	"github.com/RocketChat/rocketchat-tui/cache"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// It is used to login using email and password to rest and realtime api client throw err if not able to login.
// After successfull login rest and realtime client is set into state, to be used in all other apis call using them.
// User token and token expiration date is added in cache database to use them for future login.
// LoginScreen state is changed and terminal will show TUI.
// All subscribed channels, groups and DMs are fetched.
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

// It is used to login using token stored in cache database.
// After successfull login rest and realtime client is set into state, to be used in all other apis call using them.
// LoginScreen state is changed and terminal will show TUI.
// All subscribed channels, groups and DMs are fetched.
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

// Initiate the login workflow, verify validation and expiration of token from cache
// If token is valid login is perform using it else user redirected to login using email and password
// After login all other details like channels, available slash commands and channel members are fetched.
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
			log.Println("Error occurred while login from token", err)
			cache.CreateUpdateCacheEntry("token", "")
			cache.CreateUpdateCacheEntry("tokenExpires", "")
			m.loginScreen.loginScreenState = "showLoginScreen"
			m.loginScreen.loggedIn = false
			m.loginScreen.err = err
			return nil
		}
		channelCmd := m.setChannelsInUiList()
		m.typing = true
		m.changeSelectedChannel(0)
		setSlashCommandsList := m.setSlashCommandsList()
		channelMembersSetCmnd := m.setChannelMembersList()
		return tea.Batch(channelCmd, setSlashCommandsList, channelMembersSetCmnd)

	} else {
		cache.CreateUpdateCacheEntry("token", "")
		cache.CreateUpdateCacheEntry("tokenExpires", "")
		m.loginScreen.loginScreenState = "showLoginScreen"
		m.loginScreen.loggedIn = false
		m.loginScreen.err = generateError("token expired, Please login again")
		return nil
	}

}

// It handles the updatation of login screen
// Key bindings for login screen are handled
// After login all other details like channels, available slash commands and channel members are fetched.
func (m *Model) handleLoginScreenUpdate(msg tea.Msg) (tea.Cmd, error) {
	if !m.loginScreen.loggedIn && m.loginScreen.loginScreenState == "showLoginScreen" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return tea.Quit, nil
			case "tab", "ctrl+down":
				if m.loginScreen.activeElement < 3 {
					m.loginScreen.activeElement = m.loginScreen.activeElement + 1
				} else {
					m.loginScreen.activeElement = 1
				}
				return nil, nil
			case "enter":
				if m.email != "" && m.password != "" {
					err := m.connectFromEmailAndPassword()
					if err != nil {
						log.Println("Error occurred while login from email", err)
						return nil, err
					}
					var cmds []tea.Cmd
					channelCmd := m.setChannelsInUiList()
					cmd := m.waitForIncomingMessage(m.msgChannel)
					m.loginScreen.loggedIn = true
					m.changeSelectedChannel(0)
					m.typing = true
					setSlashCommandsList := m.setSlashCommandsList()
					channelMembersSetCmnd := m.setChannelMembersList()
					cmds = append(cmds, channelCmd, textinput.Blink, setSlashCommandsList, channelMembersSetCmnd, cmd)
					return tea.Batch(cmds...), nil
				}
				err := generateError("Please enter email and password")
				return nil, err
			}
		}

		if m.loginScreen.activeElement == 1 {
			var cmd tea.Cmd
			m.loginScreen.emailInput, cmd = m.loginScreen.emailInput.Update(msg)
			m.email = m.loginScreen.emailInput.Value()
			return cmd, nil
		}

		if m.loginScreen.activeElement == 2 {
			var cmd tea.Cmd
			m.loginScreen.passwordInput, cmd = m.loginScreen.passwordInput.Update(msg)
			m.password = m.loginScreen.passwordInput.Value()
			return cmd, nil
		}
	}
	return nil, nil
}

// Handle user logout and restoration of the TUI state to the intial state once user is logged out
// Clear user token and its expiration date from cache
func (m *Model) handleUserLogOut() (tea.Model, tea.Cmd) {
	sUrl := m.serverUrl
	m.loginScreen.passwordInput.Reset()
	m.loginScreen.emailInput.Reset()
	m = IntialModelState(sUrl)
	cache.CreateUpdateCacheEntry("token", "")
	cache.CreateUpdateCacheEntry("tokenExpires", "")
	return m, nil
}
