# Rocket.Chat Terminal User Interface

Terminal User Interface for [Rocket.Chat](https://github.com/RocketChat) made using [Bubbletea](https://github.com/charmbracelet/bubbletea)

### Quick Start

Prerequisites:

- [Git](http://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [Golang](https://go.dev/)

## Development
- Clone this Project Repo and set up [Rocket.chat Meteor Application](https://github.com/RocketChat/Rocket.Chat) in your machine.
- Run Rocketchat Meteor Server on your `http://localhost:3000` and login/signup into a new account save your credentials for signing in TUI.
- In the RocketChat TUI root folder run `go get` in terminal to get all golang packages we are using
- Make a `.env` file in the project root directory and add below code in it.

    ```
    DEV_SERVER_URL=http://localhost:3000
    ```
- Now in the RocketChat TUI root folder run `go run main.go -debug` to run the TUI.
- we have to pass `-debug` flag so that it logs log statements in `debug.log` file.
- Hopefully you will see the TUI running.
- Enter your email and password. Press Enter.

## Structure of the Project
- The starting file of the project is `main.go`. It starts the bubbletea Program to run TUI.
- TUI models, view and controllers are present in ui folder.
- In ui folder `model.go` contain global state of TUI and methods required by bubbletea to initialise, Update and Render the TUI in terminal.
- In ui folder `view.go` contain UI code of the TUI which uses styles defined in `styles` package. We are using [lipgloss](https://github.com/charmbracelet/lipgloss) for styling the TUI.
- All Key bindings used in TUI is present in `keyBindings` package to keep them seperate from TUI so that new key bindings can be easily added when needed.
- Caching related functions are present in `cache` package