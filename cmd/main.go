package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/louislaugier/twitter-bot/internal/email"
)

func main() {
	godotenv.Load(fmt.Sprintf("../%s.env", os.Getenv("service")))

	// scraper, err := config.Login()
	// if err != nil {
	// 	log.Println(err)
	// }

	// services.InitAutofollow(scraper)
	// services.InitAutounfollow(scraper)

	email.GetValidEmailsFromCSVIntoNewCSV("input.csv", "output.csv")
}
