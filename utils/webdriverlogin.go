package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/simonedegiacomi/gphotosuploader/auth"
	"github.com/tebeka/selenium"
)

// Start a wizard that open a browser to let the user authenticate and return an auth.Credentials implementation
func StartWebDriverCookieCredentialsWizard() (*auth.CookieCredentials, error) {
	fmt.Print("\n-- WebDriver CookieCredentials Wizard --\n")
	browserName, driverAddress := askBrowserAndDriverAddress()

	// Connect to the WebDriver
	capabilities := selenium.Capabilities{
		"browserName": browserName,
	}
	webDriver, err := selenium.NewRemote(capabilities, driverAddress)
	if err != nil {
		return nil, fmt.Errorf("can't initialize selenium library (%v)", err)
	}
	defer webDriver.Close()

	if err := instructUserAndWaitForLogin(webDriver); err != nil {
		return nil, err
	}

	cookies, err := extractCookies(webDriver)
	if err != nil {
		return nil, err
	}

	// Create auth container
	credentials := auth.NewCookieCredentials(cookies, &auth.PersistentParameters{})

	credentials.PersistentParameters.UserId, err = extractUserId(webDriver)
	return credentials, err
}

func askBrowserAndDriverAddress() (string, string) {
	var browserName string
	var driverAddress string

	fmt.Print("Please insert the name of the browser to use: ")
	fmt.Scanln(&browserName)

	fmt.Println("Insert the address of the WebDriver (example: http://localhost:9515): ")
	fmt.Scanln(&driverAddress)

	return browserName, driverAddress
}

func instructUserAndWaitForLogin(webDriver selenium.WebDriver) error {
	// Navigate to the Google Photos login page
	if err := webDriver.Get(auth.LoginUrl); err != nil {
		return fmt.Errorf("can't navigate to login page (%v)", err)
	}

	// Wait for the user to reach Google Photos Homepage
	fmt.Println("\nA browser window should now apper with the Google Photos Login page.")
	fmt.Println("Once you will be redirected to the Google Photos Homepage the browser will clouse automatically.")
	fmt.Println("Please fill the form and login now")
	loginCompleted := false
	for !loginCompleted {
		time.Sleep(1 * time.Second)

		url, _ := webDriver.CurrentURL()
		if url == auth.HomeUrl {
			loginCompleted = true
		}
	}
	fmt.Println("You should now be authenticated in the browser, now I'll try to get the cookies ...")
	return nil
}

func extractCookies(webDriver selenium.WebDriver) ([]*http.Cookie, error) {
	// Get cookies from browser
	seleniumCookies, err := webDriver.GetCookies()
	if err != nil {
		return nil, fmt.Errorf("can't get cookies from WebDriver (%v)", err)
	}

	// Convert selenium cookies to go cookies
	return SeleniumToGoCookies(seleniumCookies), nil
}

// Utility to convert selenium cookies slice to go http.Cookie slice
func SeleniumToGoCookies(seleniumCookies []selenium.Cookie) []*http.Cookie {
	goCookies := []*http.Cookie{}
	for _, cookie := range seleniumCookies {
		goCookies = append(goCookies, SeleniumToGoCookie(cookie))
	}
	return goCookies
}

// utility to convert a selenium cookie to a go http.Cookie
func SeleniumToGoCookie(seleniumCookie selenium.Cookie) *http.Cookie {
	return &http.Cookie{
		Name:   seleniumCookie.Name,
		Domain: seleniumCookie.Domain,
		Path:   seleniumCookie.Path,
		Secure: seleniumCookie.Secure,
		Value:  seleniumCookie.Value,
	}
}

func extractUserId(webDriver selenium.WebDriver) (string, error) {
	// Get the user id
	res, err := webDriver.ExecuteScript(`return { id: window.WIZ_global_data.S06Grb } `, nil)
	if err != nil {
		return "", fmt.Errorf("can't get user id (%v)", err)
	}

	info := res.(map[string]interface{})
	infoID, ok1 := info["id"]
	if !ok1 {
		return "", fmt.Errorf("can't find key 'id' in info")
	}
	infoIDString, ok2 := infoID.(string)
	if !ok2 {
		return "", fmt.Errorf("can't cast infoID to string")
	}

	return infoIDString, nil
}
