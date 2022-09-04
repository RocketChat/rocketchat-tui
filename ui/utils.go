package ui

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
)

// Return first letter of a string.
func getStringFirstLetter(str string) string {
	letter := "U"
	if len(str) > 0 {
		letter = strings.ToUpper(string(str[0:1]))
	}
	return letter
}

// Compare today's date time with the token expiration date to validate token.
// Take token expiration time as string in argument and then converted into value of type time.Time to compare with today's date time.
func CheckForTokenExpiration(tokenExpirationTime string) bool {
	today := time.Now()
	i, _ := strconv.ParseInt(tokenExpirationTime, 10, 64)
	tokenExpires := time.Unix(0, (i)*int64(time.Millisecond))
	return today.Before(tokenExpires)
}

// To extract sub string from a particular position in a string until the string terminates or there is a space in the string.
func stringUsernameExtractor(str string, currentPos int) string {
	length := len(str)
	chars := []rune(str)
	word := ""
	pos := currentPos
	for pos < length && string(chars[pos]) != "" && string(chars[pos]) != " " {
		word += string(chars[pos])
		pos++
	}
	return word
}

// When user selects a username from the list then add that complete username in place of the username first letters typed by user.
// Return complete string after adding username at the starting position of the username first letters typed by user.
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

func generateError(errorString string) error {
	return errors.New(errorString)
}
