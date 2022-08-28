package ui

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/joho/godotenv"
)

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func getServerUrl() string {
	DEV_SERVER_URL := goDotEnvVariable("DEV_SERVER_URL")
	return DEV_SERVER_URL
}

func getStringFirstLetter(str string) string {
	letter := "U"
	if len(str) > 0 {
		letter = strings.ToUpper(string(str[0:1]))
	}
	return letter
}

func CheckForTokenExpiration(tokenExpirationTime string) bool {
	today := time.Now()
	i, _ := strconv.ParseInt(tokenExpirationTime, 10, 64)
	tokenExpires := time.Unix(0, (i)*int64(time.Millisecond))
	return today.Before(tokenExpires)
}

func stringUsernameExtractor(str string, currentPos int) string {
	length := len(str)
	chars := []rune(str)
	word := ""
	pos := currentPos
	for pos < length {
		if string(chars[pos]) != "" && string(chars[pos]) != " " {
			word += string(chars[pos])
		} else {
			break
		}
		pos++
	}
	return word
}

func usernameAutoCompleteString(str string, replacement string, index int, jump int) string {
	return str[:index] + replacement + str[index+jump:]
}

func doesUserExistInChannel(channelMembers []models.User, username string) int {
	for i := range channelMembers {
		if channelMembers[i].UserName == username {
			return 1
		}
	}
	return -1
}