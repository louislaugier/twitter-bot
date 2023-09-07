package services

import (
	"encoding/json"
	"net/http"

	"github.com/louislaugier/twitter-bot/src/models"
	"github.com/louislaugier/twitter-bot/src/scraper"
)

func getConnectTabUserIDs(s *scraper.Scraper) ([]string, error) {
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/kfbHAW8uW0F3PboJu827wg/ConnectTabTimeline?variables=%7B%22count%22%3A20%2C%22context%22%3A%22%7B%7D%22%7D&features=%7B%22responsive_web_graphql_exclude_directive_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22creator_subscriptions_tweet_preview_api_enabled%22%3Atrue%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22responsive_web_graphql_skip_user_profile_image_extensions_enabled%22%3Afalse%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22view_counts_everywhere_api_enabled%22%3Atrue%2C%22longform_notetweets_consumption_enabled%22%3Atrue%2C%22responsive_web_twitter_article_tweet_consumption_enabled%22%3Afalse%2C%22tweet_awards_web_tipping_enabled%22%3Afalse%2C%22freedom_of_speech_not_reach_fetch_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Atrue%2C%22longform_notetweets_rich_text_read_enabled%22%3Atrue%2C%22longform_notetweets_inline_media_enabled%22%3Atrue%2C%22responsive_web_media_download_video_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Afalse%7D", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.RequestAPI(req, nil)
	if err != nil {
		return nil, err
	}

	response := models.Response{}
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

// func getUserID(s *scraper.Scraper, username string) (*string, error) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/users/show.json?screen_name=%s", username), nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := s.RequestAPI(req, "data")
// 	if err != nil {
// 		return nil, err
// 	}

// 	r := string(res)

// 	return &r, err
// }
