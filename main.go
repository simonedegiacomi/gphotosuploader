package main

import (
	"flag"
	"os"
	"path/filepath"
	"github.com/gphotosuploader/auth"
	"github.com/fsnotify/fsnotify"
	"log"
	"fmt"
	"github.com/gphotosuploader/utils"
	"github.com/gphotosuploader/api"
	"bufio"
	"os/signal"
	"syscall"
	"strings"
)
var (
	// CLI arguments
	authFile string
	filesToUpload utils.FilesToUpload
	directoriesToWatch utils.DirectoriesToWatch
	uploadedListFile string
	watchRecursively bool
	maxConcurrentUploads int

	// Uploader
	uploader *utils.ConcurrentUploader

	// Statistic
	uploadedFilesCount = 0
	ignoredCount = 0
	errorsCount = 0
)

// Parse CLI arguments
func initCliArguments() {
	flag.StringVar(&authFile, "auth", "auth.json", "Authentication json file")
	flag.Var(&filesToUpload, "upload", "File or directory to upload")
	flag.StringVar(&uploadedListFile, "uploadedList", "uploaded.txt", "List to already uploaded files")
	flag.IntVar(&maxConcurrentUploads, "maxConcurrent", 1, "Number of max concurrent uploads")
	flag.Var(&directoriesToWatch, "watch", "Directory to watch")
	flag.BoolVar(&watchRecursively, "watchRecursively", true, "Start watching new directories in currently watched directories")

	flag.Parse()
}

// Visitor function used by filepath.Walk that when visit a file upload it
func visitAndEnqueue(path string, file os.FileInfo, err error) error {
	if !file.IsDir() {
		options := api.UploadOptions{
			FileToUpload: path,
		}
		uploader.EnqueueUpload(&options)

	}

	return nil
}

// Upload all the file and directories passed as arguments, calling filepath.Walk on each name
func uploadArgumentsFiles() {
	for _, name := range filesToUpload {
		filepath.Walk(name, visitAndEnqueue)
	}
}

func handleUploaderEvents(exiting chan bool) {
	for {
		select {
		case info := <-uploader.CompletedUploads:
			uploadedFilesCount++
			log.Printf("Upload of '%v' completed\n", info.Name)

			// Update the upload completed file
			if file, err := os.OpenFile(uploadedListFile, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0666); err != nil {
				log.Println("Can't update the uploaded file list")
			} else {
				file.WriteString(fmt.Sprintf("%v\n", info.FileToUpload))
				file.Close()
			}

		case info := <-uploader.IgnoredUploads:
			ignoredCount++
			log.Printf("Not uploading '%v', it's already been uploaded or it's not a image/video!\n", info.FileToUpload)

		case err := <-uploader.Errors:
			log.Printf("Upload error: %v\n", err)
			errorsCount++

		case <-exiting:
			exiting <- true
			break
		}
	}
}

func handleFileSystemEvents(fsWatcher *fsnotify.Watcher, exiting chan bool) {
	for {
		select {
		case event := <-fsWatcher.Events:
			if event.Op != fsnotify.Remove {
				if info, err := os.Stat(event.Name); err != nil {
					log.Println(err)
				} else {

					// Upload the content of the new file
					filepath.Walk(event.Name, visitAndEnqueue)

					// Start watching a new directory if needed
					if info.IsDir() && watchRecursively {
						fsWatcher.Add(event.Name)
					}
				}

			}

		case err := <-fsWatcher.Errors:
			log.Println(err)
		case <-exiting:
			exiting <- true
			break
		}
	}
}

func notifyUploaderOfAlreadyUploadedFiles() {
	file, err := os.OpenFile(uploadedListFile, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		uploader.AddUploadedFiles(scanner.Text())
	}
}

func initAuthentication () auth.Credentials{
	// Load authentication parameters
	credentials, err := auth.NewCookieCredentialsFromFile(authFile)
	if err != nil {
		log.Printf("Can't use '%v' as auth file\n", authFile)
		credentials = nil
	} else {
		log.Println("Auth file loaded, checking validity ...")
		validity, err := credentials.TestCredentials()
		if err != nil {
			log.Fatalf("Can't check validity of credentials (%v)\n", err)
			credentials = nil
		} else if !validity.Valid {
			log.Printf("Credentials are not valid! %v\n", validity.Reason)
			credentials = nil
		} else {
			log.Println("Auth file seems to be valid")
		}
	}

	if credentials == nil {
		fmt.Println("The uploader can't continue without valid authentication tokens ...")
		fmt.Println("Would you like to run the WebDriver CookieCredentials Wizard ? [Yes/No]")
		fmt.Println("(If you don't know what it is, refer to the README)")

		var answer string
		fmt.Scanln(&answer)
		startWizard := len(answer) > 0 && strings.ToLower(answer)[0] == 'y'

		if !startWizard {
			log.Fatalln("It's not possible to continue, sorry!")
		} else {
			credentials, err = utils.StartWebDriverCookieCredentialsWizard()
			if err != nil {
				log.Fatalf("Can't complete the login wizard, got: %v\n", err)
			} else {
				// TODO: Handle error
				credentials.SerializeToFile(authFile)
			}
		}
	}

	// Get a new At token
	log.Println("Getting a new At token ...")
	token, err := api.NewAtTokenScraper(credentials).ScrapeNewToken()
	if err != nil {
		log.Fatalf("Can't scrape a new At token (%v)\n", err)
	}
	credentials.GetRuntimeParameters().AtToken = token
	log.Println("At token taken")

	return credentials
}

func main() {

	// Parse console arguments
	initCliArguments()

	// Initialize authentication
	credentials := initAuthentication()

	// Create the uploader
	var err error
	uploader, err = utils.NewUploader(credentials, maxConcurrentUploads)
	if err != nil {
		log.Fatalf("Can't create uploader: %v\n", err)
	}

	stopHandler := make(chan bool)
	go handleUploaderEvents(stopHandler)

	// Load the list of already uploaded files
	notifyUploaderOfAlreadyUploadedFiles()

	// Upload files passed as arguments
	uploadArgumentsFiles()

	// Wait until all the uploads are completed
	uploader.WaitUploadsCompleted()

	// Start to watch all the directories if needed
	if len(directoriesToWatch) > 0 {
		// Create the watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}
		defer watcher.Close()
		go handleFileSystemEvents(watcher, stopHandler)


		// Add all the directories passed as argument to the watcher
		for _, name := range directoriesToWatch {
			if err := watcher.Add(name); err != nil {
				panic(err)
			}
		}

		log.Println("Watching 👀\nPress CTRL + C to stop")

		// Wait for CTRL + C
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
	}

	stopHandler <- true
	<-stopHandler
	stopHandler <- true
	<-stopHandler

	log.Printf("Done (%v files uploaded, %v files ignored, %v errors)", uploadedFilesCount, ignoredCount, errorsCount)
	os.Exit(0)
}