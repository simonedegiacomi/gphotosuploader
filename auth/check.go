package auth

import (
	"net/http"
)

const (
	LoginUrl = "https://photos.google.com/login"
	HomeUrl  = "https://photos.google.com/"
)

// Result of a credentials test
type CredentialsTestResult struct {
	// False if the cookies are not valid anymore
	Valid bool

	// Reason to explain a negative Valid field value
	Reason string
}

// Test the CookieCredentials object to see if the authentication cookies are valid.
// Note that this method can return false-positive results, but if it return a CredentialsTestResult with false as Valid
// the cookies are not valid for sure.
// An eventual as second return parameter try to explain why we can't determine the credentials validity
func (c *CookieCredentials) CheckCredentials() (*CredentialsTestResult, error) {
	// To check if the cookies are valid, make a request to the Google Photos Login and check if we're redirected
	res, err := c.sendLoginRequest()
	if err != nil {
		return nil, err
	}

	if res.Request.URL.String() != HomeUrl {
		return &CredentialsTestResult{
			Valid:  false,
			Reason: "Google didn't redirect us to the Photos Homepage while accessing the Login page",
		}, nil
	}

	// All seems all right
	return &CredentialsTestResult{
		Valid: true,
	}, nil
}

func (c *CookieCredentials) sendLoginRequest() (*http.Response, error) {
	if req, err := http.NewRequest("GET", LoginUrl, nil); err != nil {
		return nil, err
	} else {
		return c.Client.Do(req)
	}
}
