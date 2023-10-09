package services

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/louislaugier/twitter-bot/config"
	"github.com/louislaugier/twitter-bot/internal/scraper"
)

// InitAutofollow export
func InitAutofollow() error {
	scraper, err := config.Login()
	if err != nil {
		return err
	}

	accountIDToGetFollowersFrom, err := getUserID(scraper, os.Getenv("TWITTER_HANDLE_TO_FOLLOW_FOLLOWERS_FROM"))
	if err != nil {
		return err
	}
	selfUserID, err := getUserID(scraper, os.Getenv("TWITTER_HANDLE"))
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
				continue
			}

			isFollower, err := isFollower(scraper, selfUserID, v)
			if err != nil {
				continue
			}

			if !isFollowingOrPending && err == nil && !isFollower {
				err = followUser(scraper, v)

				if err != nil {
					log.Printf("Failed to follow %s: %s", v, err)
					continue
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

// InitAutofollow export
func InitAutounfollow() error {
	scraper, err := config.Login()
	if err != nil {
		return err
	}

	accountIDToGetFollowersFrom, err := getUserID(scraper, "freelancechain")
	if err != nil {
		return err
	}

	nextCursor := "-1"

	for {
		IDs, newNextCursor, err := getFollowing(scraper, accountIDToGetFollowersFrom, nextCursor)
		if err != nil {
			return err
		}

		for _, v := range IDs {
			err = unfollowUser(scraper, v)

			if err != nil {
				log.Printf("Failed to unfollow %s: %s", v, err)
				continue
			}
		}

		if newNextCursor == "0" {
			break
		}

		nextCursor = newNextCursor
	}

	return InitAutounfollow()
}

func followUser(s *scraper.Scraper, userID string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.twitter.com/1.1/friendships/create.json?user_id=%s", userID), nil)
	if err != nil {
		return err
	}

	_, err = s.RequestAPI(req, nil)
	if err != nil {
		if strings.Contains(err.Error(), "unable to follow more people at this time") {
			log.Println(err)
			time.Sleep(time.Minute * 30)

			return followUser(s, userID)
		} else {
			return err
		}
	}

	log.Printf("Followed %s successfully", userID)

	return nil
}

func unfollowUser(s *scraper.Scraper, userID string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.twitter.com/1.1/friendships/destroy.json?user_id=%s", userID), nil)
	if err != nil {
		return err
	}

	_, err = s.RequestAPI(req, nil)
	if err != nil {
		return err
	}

	log.Printf("Unfollowed %s successfully", userID)

	return nil
}
