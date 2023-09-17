package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/louislaugier/twitter-bot/internal/scraper"
)

func getFollowers(s *scraper.Scraper, userID string, cursor string) ([]string, string, error) {
	url := fmt.Sprintf("https://api.twitter.com/1.1/followers/ids.json?user_id=%s&cursor=%s&count=5", userID, cursor)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := s.RequestAPI(req, nil)
	if err != nil {
		return nil, "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, "", err
	}

	var followers []string
	if ids, ok := result["ids"].([]interface{}); ok {
		for _, id := range ids {
			if idFloat, ok := id.(float64); ok {
				idStr := strconv.FormatFloat(idFloat, 'f', 0, 64)
				followers = append(followers, idStr)
			}
		}
	}

	nextCursor := ""
	if nextCursorStr, ok := result["next_cursor_str"].(string); ok {
		nextCursor = nextCursorStr
	}

	return followers, nextCursor, nil
}

func isFollowingOrPending(s *scraper.Scraper, userID string) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/friendships/show.json?target_id=%s", userID), nil)
	if err != nil {
		return false, err
	}

	resp, err := s.RequestAPI(req, nil)
	if err != nil {
		return false, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return false, err
	}

	if relationship, ok := result["relationship"].(map[string]interface{}); ok {
		if source, ok := relationship["source"].(map[string]interface{}); ok {
			if following, _ := source["following"].(bool); following {
				return true, nil
			}
			if followRequestSent, _ := source["following_requested"].(bool); followRequestSent {
				return true, nil
			}
		}
	}

	return false, nil
}

func getUserID(s *scraper.Scraper, username string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/users/show.json?screen_name=%s", username), nil)
	if err != nil {
		return "", err
	}

	resp, err := s.RequestAPI(req, nil)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	if id, ok := result["id_str"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("Failed to parse response")
}

func getUsername(s *scraper.Scraper, userID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/users/show.json?user_id=%s", userID), nil)
	if err != nil {
		return "", err
	}

	resp, err := s.RequestAPI(req, nil)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	if username, ok := result["screen_name"].(string); ok {
		return username, nil
	}

	return "", fmt.Errorf("Failed to parse response")
}

// Get user IDs from "Connect" tab on Twitter webapp sidebar
func getConnectTabUserIDs(s *scraper.Scraper) ([]string, error) {
	type userResults struct {
		Result struct {
			ID     string `json:"id"`
			RestID string `json:"rest_id"`
		} `json:"result"`
	}

	type item struct {
		ItemContent struct {
			ItemType    string      `json:"itemType"`
			UserResults userResults `json:"user_results"`
		} `json:"itemContent"`
	}

	type rsp struct {
		Data struct {
			ConnectTabTimeline struct {
				Timeline struct {
					Instructions []struct {
						Type    string `json:"type"`
						Entries []struct {
							EntryID string `json:"entryId"`
							Content struct {
								EntryType string `json:"entryType"`
								Items     []struct {
									EntryID string `json:"entryId"`
									Item    item   `json:"item"`
								} `json:"items"`
							} `json:"content"`
						} `json:"entries"`
					} `json:"instructions"`
				} `json:"timeline"`
			} `json:"connect_tab_timeline"`
		} `json:"data"`
	}

	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/kfbHAW8uW0F3PboJu827wg/ConnectTabTimeline?variables=%7B%22count%22%3A20%2C%22context%22%3A%22%7B%7D%22%7D&features=%7B%22responsive_web_graphql_exclude_directive_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22creator_subscriptions_tweet_preview_api_enabled%22%3Atrue%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22responsive_web_graphql_skip_user_profile_image_extensions_enabled%22%3Afalse%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22view_counts_everywhere_api_enabled%22%3Atrue%2C%22longform_notetweets_consumption_enabled%22%3Atrue%2C%22responsive_web_twitter_article_tweet_consumption_enabled%22%3Afalse%2C%22tweet_awards_web_tipping_enabled%22%3Afalse%2C%22freedom_of_speech_not_reach_fetch_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Atrue%2C%22longform_notetweets_rich_text_read_enabled%22%3Atrue%2C%22longform_notetweets_inline_media_enabled%22%3Atrue%2C%22responsive_web_media_download_video_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Afalse%7D", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.RequestAPI(req, nil)
	if err != nil {
		return nil, err
	}

	response := rsp{}
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	userIDs := []string{}

	for _, instruction := range response.Data.ConnectTabTimeline.Timeline.Instructions {
		if instruction.Type == "TimelineAddEntries" {
			for _, entry := range instruction.Entries {
				for _, item := range entry.Content.Items {
					if item.Item.ItemContent.ItemType == "TimelineUser" {
						restID := item.Item.ItemContent.UserResults.Result.RestID
						userIDs = append(userIDs, restID)
					}
				}
			}
		}
	}

	return userIDs, nil
}
