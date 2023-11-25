package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(fmt.Sprintf("../%s.env", os.Getenv("service")))

	// scraper, err := config.Login()
	// if err != nil {
	// 	log.Println(err)
	// }

	file, err := os.Open("../internal/email/free-email-database-of-india.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	var allEmails []string
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		for _, email := range strings.Split(record[0], ";") {
			email = strings.TrimSpace(email)
			if email != "" {
				allEmails = append(allEmails, email)
			}
		}
	}
	fmt.Println("All Emails:", allEmails)

	// services.InitAutofollow(scraper)
	// services.InitAutounfollow(scraper)
}
