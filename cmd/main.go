package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/louislaugier/twitter-bot/config"
	services "github.com/louislaugier/twitter-bot/internal"
)

func main() {
	godotenv.Load(fmt.Sprintf("../%s.env", os.Getenv("service")))

	scraper, err := config.Login()
	if err != nil {
		log.Println(err)
	}

	// services.InitAutoDM(scraper)

	// services.InitAutounfollow()

	services.InitAutofollow(scraper)
}
