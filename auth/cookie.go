package auth

import (
	"net/http"
	"os"
	"net/http/cookiejar"
	"encoding/json"
	"net/url"
	"fmt"
)

type CookieCredentials struct {
	client	*http.Client
}

func NewCookieCredentials () *CookieCredentials {
	return NewCookieCredentialsFromFile(nil)
}

func NewCookieCredentialsFromFile (file *os.File) *CookieCredentials {
	cookieJar, _ := cookiejar.New(nil)

	cookiesUrl, err := url.Parse("https://photos.google.com")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if file != nil {
		cookies := []*http.Cookie{}
		json.NewDecoder(file).Decode(&cookies)
		cookieJar.SetCookies(cookiesUrl, cookies)
	}

	return &CookieCredentials{
		client: &http.Client{
			Jar: cookieJar,
		} ,
	}
}

func (c *CookieCredentials) GetClient () *http.Client {
	return c.client
}