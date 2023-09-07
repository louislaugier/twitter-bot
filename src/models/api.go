package models

// Response export
type Response struct {
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

type item struct {
	ItemContent struct {
		ItemType    string      `json:"itemType"`
		UserResults userResults `json:"user_results"`
	} `json:"itemContent"`
}

type userResults struct {
	Result struct {
		ID     string `json:"id"`
		RestID string `json:"rest_id"`
	} `json:"result"`
}
