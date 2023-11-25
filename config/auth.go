package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/louislaugier/twitter-bot/internal/scraper"
)

// Login export
func Login() (*scraper.Scraper, error) {
	scraper := scraper.New()

	loadCookies(scraper)

	if !scraper.IsLoggedIn() {
		err := scraper.Login(os.Getenv("TWITTER_HANDLE"), os.Getenv("TWITTER_PWD"))
		if err != nil {
			return nil, err
		}

		saveCookies(scraper)
	}

	return scraper, nil
}

func saveCookies(scraper *scraper.Scraper) {
	cookies := scraper.GetCookies()

	f, _ := os.Create(fmt.Sprintf("../%s.json", os.Getenv("service")))
	js, _ := json.Marshal(cookies)

	f.Write(js)
}

func loadCookies(scraper *scraper.Scraper) {
	f, _ := os.Open(fmt.Sprintf("../%s.json", os.Getenv("service")))

	cookies := []*http.Cookie{}
	json.NewDecoder(f).Decode(&cookies)

	if cookies != nil {
		scraper.SetCookies(cookies)
	}
}
