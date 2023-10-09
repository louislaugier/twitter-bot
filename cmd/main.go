package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	services "github.com/louislaugier/twitter-bot/internal"
)

func main() {
	godotenv.Load(fmt.Sprintf("../%s.env", os.Getenv("service")))

	err := services.InitAutofollow()
	// err := services.InitAutounfollow()
	if err != nil {
		panic(err)
	}
}
