package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

const (
	cookieDomainWithProtocol = "https://photos.google.com"
	cookiesDomain            = "photos.google.com"
	googleCookiesDomain      = ".google.com"
)

type PersistentParameters struct {
	UserId string `json:"userId"`
}

type RuntimeParameters struct {
	AtToken string
}

type CookieCredentials struct {
	Client               *http.Client
	PersistentParameters *PersistentParameters
	RuntimeParameters    *RuntimeParameters
}

type AuthFile struct {
	Cookies []*http.Cookie `json:"cookies"`

	// Parameters used by the requests (user id, etc ...)
	PersistentParameters *PersistentParameters `json:"persistantParameters"`
}

// Restore an CookieCredentials object from a JSON file
func NewCookieCredentialsFromFile(fileName string) (*CookieCredentials, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("auth: Can't open %v", fileName)
	}
	defer file.Close()

	return NewCookieCredentialsFromJson(file)
}

// Restore an CookieCredentials object from a JSON
func NewCookieCredentialsFromJson(in io.Reader) (*CookieCredentials, error) {
	authFile := AuthFile{}
	if err := json.NewDecoder(in).Decode(&authFile); err != nil {
		return nil, fmt.Errorf("auth: Can't read the JSON AuthFile (%v)", err)
	}

	return NewCookieCredentials(authFile.Cookies, authFile.PersistentParameters), nil
}

func NewCookieCredentials(cookies []*http.Cookie, parameters *PersistentParameters) *CookieCredentials {
	return &CookieCredentials{
		Client: &http.Client{
			Jar: createJarWithCookies(cookies),
		},
		PersistentParameters: parameters,
		RuntimeParameters:    &RuntimeParameters{},
	}
}

func createJarWithCookies(cookies []*http.Cookie) *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	cookiesUrl, _ := url.Parse(cookieDomainWithProtocol)

	jar.SetCookies(cookiesUrl, cookies)

	return jar
}

// Serialize the CookieCredentials object into a JSON file, to be restored in the future using
// NewCookieCredentialsFromJsonFile
func (c *CookieCredentials) SerializeToFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("auth: Can't create the file %v (%v)", fileName, err)
	}
	defer file.Close()

	return c.Serialize(file)
}

// Serialize the CookieCredentials object into a JSON object, to be restored in the future using
// NewCookieCredentialsFromJson
func (c *CookieCredentials) Serialize(out io.Writer) error {
	cookiesUrl, _ := url.Parse(cookieDomainWithProtocol)
	cookies := c.Client.Jar.Cookies(cookiesUrl)

	prepareCookiesForSerialization(cookies)

	return json.NewEncoder(out).Encode(&AuthFile{
		Cookies:              cookies,
		PersistentParameters: c.PersistentParameters,
	})
}

func prepareCookiesForSerialization(cookies []*http.Cookie) {
	for _, cookie := range cookies {
		if cookie.Name == "OTZ" {
			cookie.Domain = cookiesDomain
		} else {
			cookie.Domain = googleCookiesDomain
		}
		cookie.Path = "/"
	}
}

