package utils

import (
	"github.com/tebeka/selenium"
	"time"
	"net/http"
	"github.com/gphotosuploader/auth"
	"fmt"
)


func StartWebDriverCookieCredentialsWizard () (*auth.CookieCredentials, error) {
	fmt.Print("\n-- WebDriver CookieCredentials Wizard --\n")

	// Get browser name and address
	fmt.Print("Please insert the name of the browser to use: ")
	var browserName string
	fmt.Scanln(&browserName)

	fmt.Println("Insert the address of the WebDriver (example: http://localhost:9515): ")
	var driverAddress string
	fmt.Scanln(&driverAddress)


	// Connect to the WebDriver
	capabilities := selenium.Capabilities{
		"browserName": browserName,
	}
	webDriver, err := selenium.NewRemote(capabilities, driverAddress)
	if err != nil {
		return nil, fmt.Errorf("Can't initialize selenium library (%v)", err)
	}
	defer webDriver.Close()

	// Navigate to the Google Photos login page
	if err := webDriver.Get("https://photos.google.com/login"); err != nil {
		return nil, fmt.Errorf("Can't navigate to login page (%v)", err)
	}

	// Wait for the user to reach Google Photos Homepage
	fmt.Println("\nA browser window should now apper with th eGoogle Photos Login page.")
	fmt.Println("Once you will be redirected to the Google Photos Homepage the browser will clouse automatically.")
	fmt.Println("Please fill the form and login now")

	loginCompleted := false
	for !loginCompleted {
		time.Sleep(1 * time.Second)

		url, _ := webDriver.CurrentURL()
		if url == "https://photos.google.com/" {
			loginCompleted = true
		}
	}
	fmt.Println("You should now be authenticated in the browser, now I'll try to get the cookies ...")

	// Get cookies from browser
	seleniumCookies, err := webDriver.GetCookies()
	if err != nil {
		return nil, fmt.Errorf("Can't get cookies from WebDriver (%v)", err)
	}

	// Convert selenium cookies to go cookies
	cookies := SeleniumToGoCookies(seleniumCookies)

	// Create auth container
	credentials := auth.NewCookieCredentials(cookies, &auth.PersistentParameters{})

	// Get the user id
	res, err := webDriver.ExecuteScript(`return { id: window.WIZ_global_data.S06Grb } `, nil)
	if err != nil {
		return nil, fmt.Errorf("Can't get user id (%v)", err)
	}

	info := res.(map[string]interface{})
	credentials.GetPersistentParameters().UserId = info["id"].(string)


	return credentials, nil
}



func SeleniumToGoCookies (seleniumCookies[] selenium.Cookie) []*http.Cookie {
	goCookies := []*http.Cookie{}
	for _, cookie := range seleniumCookies {
		goCookies = append(goCookies, SeleniumToGoCookie(cookie))
	}
	return goCookies
}

func SeleniumToGoCookie (seleniumCookie selenium.Cookie) *http.Cookie {
	return &http.Cookie{
		Name: seleniumCookie.Name,
		Domain: seleniumCookie.Domain,
		Path: seleniumCookie.Path,
		Secure: seleniumCookie.Secure,
		Value: seleniumCookie.Value,
	}
}