package botsky

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const baseUrl = "https://cardyb.bsky.app/v1/extract?url="

type cardybMetadata struct {
	Error       string `json:"error"`
	LikelyType  string `json:"likely_type"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

func getMetadata(u *url.URL) (metadata cardybMetadata, err error) {
	res, err := http.Get(baseUrl + url.QueryEscape(u.String()))
	if err != nil {
		return metadata, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status code:", res.StatusCode)
		return metadata, errors.New("Request failed")
	}

	// Read and unmarshal the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return metadata, err
	}

	err = json.Unmarshal(body, &metadata)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return metadata, err
	}
	return metadata, nil
}
