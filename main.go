package main

import (
	"github.com/joho/godotenv"
	"github.com/louislaugier/twitter-bot/src/services"
)

func main() {
	godotenv.Load(".env")

	err := services.InitAutofollow()
	if err != nil {
		panic(err)
	}
}
