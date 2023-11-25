package services

import (
	"bufio"
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
func InitAutofollow(sc *scraper.Scraper) error {
	scraper, err := config.Login()
	if err != nil {
		log.Println(err)
		return err
	}
	if sc != nil {
		scraper = sc
	}

	accountIDToGetFollowersFrom, err := GetUserID(scraper, os.Getenv("TWITTER_HANDLE_TO_FOLLOW_FOLLOWERS_FROM"))
	if err != nil {
		return err
	}
	selfUserID, err := GetUserID(scraper, os.Getenv("TWITTER_HANDLE"))
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
				log.Println(err)
				log.Println("Non-existing account or bot")
				continue
			}

			isFollower, err := isFollower(scraper, selfUserID, v)
			if err != nil || isFollower {
				continue
			}

			if !isFollowingOrPending && err == nil {
				err = followUser(scraper, v)

				if err != nil {
					time.Sleep(time.Minute * 15)
					err = followUser(scraper, v)
					if err != nil {
						log.Printf("Failed to follow %s: %s", v, err)
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

	return InitAutofollow(sc)
}

// InitAutofollow export
func InitAutounfollow(scraper *scraper.Scraper) error {
	accountIDToGetFollowersFrom, err := GetUserID(scraper, os.Getenv("TWITTER_HANDLE"))
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

	return InitAutounfollow(scraper)
}

// InitAutoDM export
func InitAutoDM(sc *scraper.Scraper) error {
	scraper, err := config.Login()
	if err != nil {
		return err
	}
	if sc != nil {
		scraper = sc
	}

	selfUserID, err := GetUserID(scraper, os.Getenv("TWITTER_HANDLE"))
	if err != nil {
		return err
	}

	log.Println(selfUserID)

	// Read user IDs from the file and store them in a set
	// excludedUserIDs, err := readExcludedUserIDsFromFile("../exclude-dm.txt")
	// if err != nil {
	// 	return err
	// }

	nextCursor := "-1"

	for {
		IDs, newNextCursor, err := getFollowers(scraper, selfUserID, nextCursor)
		if err != nil {
			return err
		}
		log.Println(IDs)

		for _, v := range IDs {
			// Check if the user ID is in the excluded set, if not, send the direct message
			// if _, excluded := excludedUserIDs[v]; !excluded {
			// Use the recursive function to send the direct message
			err := SendDirectMessage(scraper, v, "https://twitter.com/FreelanceChain/status/1719718728802156593")
			if err != nil {
				log.Println(err)
				log.Println("error triggered, checking if user exists")
				username, err := GetUsername(scraper, v)
				if err != nil {
					log.Println("non existing user, dismissing error & user")
					continue
				}

				log.Println("user exists:", username)
				SendDirectMessageRecursive(scraper, v, "https://twitter.com/FreelanceChain/status/1719718728802156593")
			}
			// } else {
			// 	println("excluded")
			// }
		}

		if newNextCursor == "0" {
			break
		}

		nextCursor = newNextCursor
	}

	return nil
}

// readExcludedUserIDsFromFile reads user IDs from the given file and returns them in a set
func readExcludedUserIDsFromFile(filename string) (map[string]struct{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	excludedUserIDs := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		userID := scanner.Text()
		excludedUserIDs[userID] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return excludedUserIDs, nil
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

func SendDirectMessage(s *scraper.Scraper, userID, message string) error {
	url := "https://api.twitter.com/1.1/direct_messages/events/new.json"
	body := fmt.Sprintf(`{
		"event": {
			"type": "message_create",
			"message_create": {
				"target": {
					"recipient_id": "%s"
				},
				"message_data": {
					"text": "%s"
				}
			}
		}
	}`, userID, message)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	_, err = s.RequestAPI(req, nil)
	if err != nil {
		return err
	}

	log.Printf("Sent direct message to user %s: %s", userID, message)

	return nil
}

func SendDirectMessageRecursive(scraper *scraper.Scraper, userID string, message string) error {
	err := SendDirectMessage(scraper, userID, message)
	if err != nil {
		log.Println(err)
		println("Too many requests, sleeping and retrying in 15 mins...")
		time.Sleep(time.Minute * 15)
		// Retry sending the direct message recursively
		err = SendDirectMessageRecursive(scraper, userID, message)
	}
	return err
}
