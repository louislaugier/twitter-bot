package main

import "github.com/louislaugier/twitter-bot/internal/email"

func main() {
	// service := os.Getenv("service")
	// godotenv.Load(fmt.Sprintf("../%s.env", os.Getenv("service")))

	// scraper, err := config.Login()
	// if err != nil {
	// 	log.Println(err)
	// }

	// if service == "freelancechain" {

	// 	// go services.InitAutoDM(scraper, func() *string {
	// 	// 	str := "binance"
	// 	// 	return &str
	// 	// }())
	// 	// go services.InitAutoDM(scraper, nil)

	// 	services.InitAutofollow(scraper)
	// 	// services.InitAutounfollow(scraper)
	// } else if service == "tweeter-id" {
	// 	services.InitAutofollow(scraper)
	// 	// services.InitAutounfollow(scraper)

	// }

	email.GetValidEmailsFromCSVIntoNewCSV("input.csv", "output.csv")
}
