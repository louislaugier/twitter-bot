package services

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/louislaugier/twitter-bot/src/scraper"
)

// InitAutofollow export
func InitAutofollow() error {
	scraper, err := Login()
	if err != nil {
		return err
	}

	IDs, err := getConnectTabUserIDs(scraper)
	if err != nil {
		return err
	}

	for _, v := range IDs {
		err = followUser(scraper, v)
		if err != nil {
			log.Printf("Failed to follow %s: %s", v, err)
			if strings.Contains(err.Error(), "unable to follow more people at this time") {
				time.Sleep(time.Minute * 15)
			} else {
				continue
			}
		}
	}

	return InitAutofollow()
}

func followUser(s *scraper.Scraper, userID string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.twitter.com/1.1/friendships/create.json?user_id=%s", userID), nil)
	if err != nil {
		return err
	}

	_, err = s.RequestAPI(req, nil)
	if err != nil {
		return err
	}

	log.Printf("Followed %s successfully", userID)

	return nil
}
