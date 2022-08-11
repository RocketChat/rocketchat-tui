package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	// "github.com/RocketChat/Rocket.Chat.Go.SDK/models"
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

func PrintToLogFile(v ...interface{}) {
	f, err := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(v)
}

func CheckForTokenExpiration(tokenExpirationTime string) bool {
	today := time.Now()
	i, _ := strconv.ParseInt(tokenExpirationTime, 10, 64)
	tokenExpires := time.Unix(0, (i)*int64(time.Millisecond))
	return today.Before(tokenExpires)
}
