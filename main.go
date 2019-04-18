package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/simonedegiacomi/gphotosuploader/api"
	"github.com/simonedegiacomi/gphotosuploader/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/simonedegiacomi/gphotosuploader/auth"
	"github.com/simonedegiacomi/gphotosuploader/version"
)

var (
	// CLI arguments
	authFile             string
	filesToUpload        utils.FilesToUpload
	directoriesToWatch   utils.DirectoriesToWatch
	albumId              string
	albumName            string
	uploadedListFile     string
	watchRecursively     bool
	maxConcurrentUploads int
	eventDelay           time.Duration
	printVersion         bool

	// Uploader
	uploader *utils.ConcurrentUploader
	timers   = make(map[string]*time.Timer)

	// Statistics
	uploadedFilesCount = 0
	ignoredCount       = 0
	errorsCount        = 0
)

func main() {
	parseCliArguments()
	if printVersion {
		fmt.Printf("Hash:\t%s\nCommit date:\t%s\n", version.Hash, version.Date)
		os.Exit(0)
	}

	credentials := initAuthentication()

	var err error

	// Create Album first to get albumId
	if albumName != "" {
		albumId, err = api.CreateAlbum(credentials, albumName)
		if err != nil {
			log.Fatalf("Can't create album: %v\n", err)
		}
		log.Printf("New album with ID '%v' created\n", albumId)
	}

	uploader, err = utils.NewUploader(credentials, albumId, maxConcurrentUploads)
	if err != nil {
		log.Fatalf("Can't create uploader: %v\n", err)
	}

	stopHandler := make(chan bool)
	go handleUploaderEvents(stopHandler)

	loadAlreadyUploadedFiles()

	// Upload files passed as arguments
	uploadArgumentsFiles()

	// Wait until all the uploads are completed
	uploader.WaitUploadsCompleted()

	// Start to watch all the directories if needed
	if len(directoriesToWatch) > 0 {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}
		defer watcher.Close()
		go handleFileSystemEvents(watcher, stopHandler)

		// Add all the directories passed as argument to the watcher
		for _, name := range directoriesToWatch {
			if err := startToWatch(name, watcher); err != nil {
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

// Parse CLI arguments
func parseCliArguments() {
	flag.StringVar(&authFile, "auth", "auth.json", "Authentication json file")
	flag.Var(&filesToUpload, "upload", "File or directory to upload")
	flag.StringVar(&albumId, "album", "", "Use this parameter to move new images to a specific album")
	flag.StringVar(&albumName, "albumName", "", "Use this parameter to move new images to a new album")
	flag.StringVar(&uploadedListFile, "uploadedList", "uploaded.txt", "List to already uploaded files")
	flag.IntVar(&maxConcurrentUploads, "maxConcurrent", 1, "Number of max concurrent uploads")
	flag.Var(&directoriesToWatch, "watch", "Directory to watch")
	flag.BoolVar(&watchRecursively, "watchRecursively", true, "Start watching new directories in currently watched directories")
	delay := flag.Int("eventDelay", 3, "Distance of time to wait to consume different events of the same file (seconds)")
	flag.BoolVar(&printVersion, "version", false, "Print version and commit date")

	flag.Parse()

	// Check flags
	if albumId != "" && albumName != "" {
		log.Fatalf("Can't use album and albumName at the same time\n")
	}

	// Convert delay as int into duration
	eventDelay = time.Duration(*delay) * time.Second
}

func initAuthentication() auth.CookieCredentials {
	// Load authentication parameters
	credentials, err := auth.NewCookieCredentialsFromFile(authFile)
	if err != nil {
		log.Printf("Can't use '%v' as auth file\n", authFile)
		credentials = nil
	} else {
		log.Println("Auth file loaded, checking validity ...")
		validity, err := credentials.CheckCredentials()
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
	token, err := api.NewAtTokenScraper(*credentials).ScrapeNewAtToken()
	if err != nil {
		log.Fatalf("Can't scrape a new At token (%v)\n", err)
	}
	credentials.RuntimeParameters.AtToken = token
	log.Println("At token taken")

	return *credentials
}

// Upload all the file and directories passed as arguments, calling filepath.Walk on each name
func uploadArgumentsFiles() {
	for _, name := range filesToUpload {
		filepath.Walk(name, func(path string, file os.FileInfo, err error) error {
			if !file.IsDir() {
				uploader.EnqueueUpload(path)
			}

			return nil
		})
	}
}

func handleUploaderEvents(exiting chan bool) {
	for {
		select {
		case info := <-uploader.CompletedUploads:
			uploadedFilesCount++
			log.Printf("Upload of '%v' completed\n", info)

			// Update the upload completed file
			if file, err := os.OpenFile(uploadedListFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
				log.Println("Can't update the uploaded file list")
			} else {
				file.WriteString(info + "\n")
				file.Close()
			}

		case info := <-uploader.IgnoredUploads:
			ignoredCount++
			log.Printf("Not uploading '%v', it's already been uploaded or it's not a image/video!\n", info)

		case err := <-uploader.Errors:
			log.Printf("Upload error: %v\n", err)
			errorsCount++

		case <-exiting:
			exiting <- true
			break
		}
	}
}

func startToWatch(filePath string, fsWatcher *fsnotify.Watcher) error {
	if watchRecursively {
		return filepath.Walk(filePath, func(path string, file os.FileInfo, err error) error {
			if file.IsDir() {
				return fsWatcher.Add(path)
			}
			return nil
		})
	} else {
		return fsWatcher.Add(filePath)
	}
}

func handleFileChange(event fsnotify.Event, fsWatcher *fsnotify.Watcher) {
	// Use a map of timer to ignore different consecutive events for the same file.
	// (when the os writes a file to the disk, sometimes it repetitively sends same events)
	if timer, exists := timers[event.Name]; exists {

		// Cancel the timer
		cancelled := timer.Stop()

		if cancelled && event.Op != fsnotify.Remove && event.Op != fsnotify.Rename {
			// Postpone the file upload
			timer.Reset(eventDelay)
		}
	} else if event.Op != fsnotify.Remove && event.Op != fsnotify.Rename {
		timer = time.AfterFunc(eventDelay, func() {
			log.Printf("Finally consuming events for the %v file", event.Name)

			if info, err := os.Stat(event.Name); err != nil {
				log.Println(err)
			} else if !info.IsDir() {

				// Upload file
				uploader.EnqueueUpload(event.Name)
			} else if watchRecursively {

				startToWatch(event.Name, fsWatcher)
			}
		})
		timers[event.Name] = timer
	}
}

func handleFileSystemEvents(fsWatcher *fsnotify.Watcher, exiting chan bool) {
	for {
		select {
		case event := <-fsWatcher.Events:
			handleFileChange(event, fsWatcher)

		case err := <-fsWatcher.Errors:
			log.Println(err)

		case <-exiting:
			exiting <- true
			return
		}
	}
}

func loadAlreadyUploadedFiles() {
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
