# G Photos Uploader - Beta
[![Build Status](https://travis-ci.org/simonedegiacomi/gphotosuploader.svg?branch=master)](https://travis-ci.org/simonedegiacomi/gphotosuploader)

## Why? What is this?
Google Photos does not have a desktop uploader for Linux, ~~neither an API to upload a photo programmatically.~~ (now there is an [Official Google Photos API](https://developers.google.com/photos/)).

G Photos Uploader lets you upload photos from Linux (and, in theory, any OS for which you can compile a Go program) specifying the file name or watching a directory for changes.
Furthermore, the project can also be used as a library that you can include in other Go programs.

## Disclaimer
G Photos Uploader is an unofficial tool, I (and any possible contributor) don't guarantee any result. Any security or other kind of issues are at your own risk.

## Requirements
To use the tool you need to install [Go](https://golang.org/) and Git (used by ```go get``` to download the dependencies). If you will use the authentication wizard you will also need a WebDriver.

## Install

```sh
go get github.com/simonedegiacomi/gphotosuploader
```

## How can I use it?
### Standalone tool
To launch the tool you have two options:
- Add the $GOPATH/bin folder to your path: doing this you can start the program just typing ```gphotosuploader```;
- Enter the project folder and use ```go run main.go```;

To use G Photos Uploader as a standalone tool you need to be authenticated. Authentication is implemented with a JSON file that contains your cookies and user ID.

#### Authentication
Every time you run the tool, it will check for the auth file. If the file is not found or the cookies seem to be expired, the tool will ask you if you want to run a wizard to get new cookies.

##### Authentication wizard
The authentication wizard uses the WebDrivers protocol, which is usually used to perform automation tests, that allows G Photos Uploader to control a browser and read the cookies from it. To use the WebDrivers Protocol you need to install a web driver (e.g. chromedriver):

- On Ubuntu
```sh
sudo apt-get install chromium-chromedriver

# Create a link to launch the driver just typing 'chromedriver'
sudo ln -s /usr/lib/chromium-browser/chromedriver /usr/bin/chromedriver

# Then launch
chromedriver
```

- On macOS
Install [Homebrew](https://brew.sh/) and then:
```sh
brew install chromedriver

# Then launch
chromedriver
```

- On Windows:
    + Download latest Chrome Web Driver [from Google](https://sites.google.com/a/chromium.org/chromedriver/downloads);
    + Copy chromedriver.exe in the path (`C:\WINDOWS` for example);
    + Then launch it with a command prompt or `Win key + R` then `chromedriver.exe`;


Note: If you are running G Photos Uploader on a headless machine, you can run chromedriver on a separate machine as such:
```sh
chromedriver --whitelisted-ips="HEADLESS_MACHINE_IP"
```

When the Driver starts it will print the address at which it is listening.
Once you enter the name of the browser (refer to browserName [here](https://github.com/SeleniumHQ/selenium/wiki/DesiredCapabilities)) and the address of the web driver in the tool, a new browser window will appear with the Google Photos Login page.
Then you can login with your account just like you always do. When you're logged in the tool will read the cookies from the browser, save them into the auth file and close the browser window.  
You can now stop the web driver server.

##### Authentication using a Chrome extension
You can also get the authentication file using a Chrome extension. You can read more about it [here](https://github.com/simonedegiacomi/gphotosuploader/tree/master/crx-auth).


#### Upload a photo or watch a directory
Once you have the auth file, you're ready to go. For example, to upload a file named image.png:
```sh
gphotosuploader --upload ./image.png
```

Or to watch a directory:
```sh
gphotosuploader --watch path/to/photos --maxConcurrent 4
```

You can even upload all the photos of a directory and then start to watch another one:
```sh
gphotosuploader --upload /path/to/old/photos --upload /downloads/cat.png --watch path/to/new/photos
```

If you also want to add your photos to a specific existing album you can use the 'album' argument:
```sh
gphotosuploader --album albumId --upload ./image.png
```
Where the albumId is the string that you see in the url when you open the album in the Google Photos Web App
(something like: https://photos.google.com/u/2/album/album_id)

The tool creates a file (default name: uploaded.txt) which is a list of uploaded files, which will not be
re-uploaded. You can specify your own file using the uploadedList argument.
To see all the available arguments, use --help.

### Library
You can read a simple example [here](documentation/examples/simple.go) or get the documentation [here](http://godoc.org/github.com/simonedegiacomi/gphotosuploader).

## Development
if you want to continue the development of this tool/library, execute first the following script:
```
githooks/create-links
```
This will links the hooks used to handle the version of the tool.

## Used libreries
* [fsnotify](https://github.com/fsnotify/fsnotify): To watch for file system events;
* [Selenium](https://github.com/tebeka/selenium): To authenticate using a browser;


## Creators:
* [simonedegiacomi](https://github.com/simonedegiacomi)
* [alessiofaieta](https://github.com/alessiofaieta)
