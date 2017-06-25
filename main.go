package main

import (
	"flag"
	"os"
	"path/filepath"
	"gphotosuploader/auth"
	"github.com/fsnotify/fsnotify"
	"log"
	"fmt"
	"gphotosuploader/utils"
	"gphotosuploader/api"
	"bufio"
	"io/ioutil"
	"strconv"
)

var (
	// CLI arguments
	cookiesFile string
	numberFile string
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
	flag.StringVar(&cookiesFile, "cookies", "cookies.json", "File with the cookies to authenticated the requests")
	flag.StringVar(&numberFile, "number", "number", "File that constains the number used to enable the image")
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

		case info := <- uploader.IgnoredUploads:
			ignoredCount++
			log.Printf("Not uploading '%v', it's already been uploaded!\n", info.FileToUpload)

		case err := <-uploader.Errors:
			log.Printf("Upload error: %v\n", err)
			errorsCount++

		case <-exiting:
			exiting <- true
			break
		}
	}
}

func handleFileSystemEvents(fsWatcher *fsnotify.Watcher) {
	select {
	case event := <-fsWatcher.Events:
		if event.Op == fsnotify.Create {
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
	}
}

func notifyUploaderOfAlreadyUploadedFiles () {
	file, err := os.OpenFile(uploadedListFile, os.O_CREATE, 0666)
	if err != nil {
		panic (err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		uploader.AddUploadedFiles(scanner.Text())
	}
}

func main() {

	// Parse console arguments
	initCliArguments()

	// Load credentials
	credentials, err := auth.NewCookieCredentialsFromFile(cookiesFile)
	if err != nil {
		panic(fmt.Sprintf("Can't use '%v' as cookies file", cookiesFile))
	}

	// Get a new API token
	token, err := api.NewTokenScraper(credentials).ScrapeNewToken()
	if err != nil {
		panic(err)
	}
	credentials.SetAPIToken(token)

	// Read the enable number
	content, err := ioutil.ReadFile(numberFile)
	if err != nil {
		panic(err)
	}
	number, err := strconv.Atoi(string(content))
	if err != nil {
		log.Panic("Can'r read number from number file")
	}
	credentials.SetEnableNumber(number)


	// Create the uploader
	uploader, err = utils.NewUploader(credentials, maxConcurrentUploads)
	if err != nil {
		panic(fmt.Sprintf("Can't create uploader: %v\n", err))
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
		go handleFileSystemEvents(watcher)


		// Add all the directories passed as argument to the watcher
		for _, name := range directoriesToWatch {
			if err := watcher.Add(name); err != nil {
				panic(err)
			}
		}

		log.Println("Watching 👀")

		// Wait indefinitely
		<-(make(chan bool))
	}

	stopHandler <- true
	<-stopHandler
	log.Printf("Done (%v files uploaded, %v files ignored, %v errors)", uploadedFilesCount, ignoredCount, errorsCount)
}