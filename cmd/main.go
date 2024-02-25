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
	service := os.Getenv("service")
	godotenv.Load(fmt.Sprintf("../%s.env", os.Getenv("service")))

	scraper, err := config.Login()
	if err != nil {
		log.Println(err)
	}

	if service == "freelancechain" {

		// go services.InitAutoDM(scraper, func() *string {
		// 	str := "binance"
		// 	return &str
		// }())
		// go services.InitAutoDM(scraper, nil)

		services.InitAutofollow(scraper)
		// services.InitAutounfollow(scraper)
	} else if service == "tweeter-id" {
		services.InitAutofollow(scraper)
		// services.InitAutounfollow(scraper)

	}
}
