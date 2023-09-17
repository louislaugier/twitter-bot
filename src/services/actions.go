package services

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/louislaugier/twitter-bot/src/services/scraper"
)

// InitAutofollow export
func InitAutofollow() error {
	scraper, err := Login()
	if err != nil {
		return err
	}

	accountIDToGetFollowersFrom, err := getUserID(scraper, os.Getenv("TWITTER_HANDLE_TO_FOLLOW_FOLLOWERS_FROM"))
	if err != nil {
		return err
	}

	nextCursor := "-1"

	for {
		IDs, newNextCursor, err := getFollowers(scraper, accountIDToGetFollowersFrom, nextCursor)
		if err != nil {
			return err
		}

		for _, v := range IDs {
			isFollowingOrPending, err := isFollowingOrPending(scraper, v)
			if err != nil {
				if strings.Contains(err.Error(), "Too Many Requests") {
					time.Sleep(time.Minute * 5)
				}
			}

			if !isFollowingOrPending && err == nil {
				err = followUser(scraper, v)

				if err != nil {
					log.Printf("Failed to follow %s: %s", v, err)
					if strings.Contains(err.Error(), "unable to follow more people at this time") {
						time.Sleep(time.Minute * 30)
					} else {
						continue
					}
				}
			}
		}

		if newNextCursor == "0" {
			break
		}

		nextCursor = newNextCursor
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
