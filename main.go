package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/louislaugier/twitter-bot/src/services"
)

func main() {
	godotenv.Load(fmt.Sprintf("%s.env", os.Getenv("service")))

	err := services.InitAutofollow()
	if err != nil {
		panic(err)
	}
}
