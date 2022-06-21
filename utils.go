package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Credentials struct {
	email string
	pass  string
}

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func getUserCredentails() *Credentials {
	EMAIL := goDotEnvVariable("EMAIL")
	PASS := goDotEnvVariable("PASS")

	data := &Credentials{
		email: EMAIL,
		pass:  PASS,
	}
	return data
}

func getServerUrl() string {
	DEV_SERVER_URL := goDotEnvVariable("DEV_SERVER_URL")
	return DEV_SERVER_URL
}
