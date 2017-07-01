package api

import (
	"github.com/simonedegiacomi/gphotosuploader/auth"
	"net/http"
	"golang.org/x/net/html"
	"strings"
	"encoding/json"
)

const (
	GooglePhotoUrl = "https://photos.google.com/"
)

type TokenScraper struct {
	credentials auth.Credentials
}

func NewAtTokenScraper(credentials auth.Credentials) *TokenScraper {
	return &TokenScraper{
		credentials: credentials,
	}
}

func (ts *TokenScraper) ScrapeNewToken() (string, error) {
	req, err := http.NewRequest("GET", GooglePhotoUrl, nil)
	if err != nil {
		return "", err
	}


	// Make the request
	res, err := ts.credentials.GetClient().Do(req)
	if err != nil {
		return "", err
	}

	t := html.NewTokenizer(res.Body)
	found := false
	var script string
	for !found {
		tt := t.Next()
		if tt == html.StartTagToken && t.Token().Data == "script" {

			// Go to the content
			t.Next()

			// Get the script string
			script = t.Token().Data

			found = true
		}
	}

	// Clean the script string
	equalsIndex := strings.Index(script, "=")
	start := equalsIndex + 1
	end := len(script) - 1
	script = script[start:end]

	// Parse the json
	object := ApiTokenContainer{}
	if err = json.NewDecoder(strings.NewReader(script)).Decode(&object); err != nil {
		panic(err)
	}

	return object.Token, nil
}