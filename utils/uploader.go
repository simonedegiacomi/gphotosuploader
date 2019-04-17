package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/GaPhi/gphotosuploader/api"
	"github.com/simonedegiacomi/gphotosuploader/auth"
)

// Simple client used to implement the tool that can upload multiple photos or videos at once
type ConcurrentUploader struct {
	credentials auth.CookieCredentials

	// Optional field to specify the destination album
	albumId string

	// Buffered channel to limit concurrent uploads
	concurrentLimiter chan bool

	// Map of uploaded files (used as a set)
	uploadedFiles map[string]bool

	// Waiting group used for the implementation of the Wait method
	waitingGroup sync.WaitGroup

	// Flag to indicate if the client is waiting for all the upload to finish
	waiting bool

	CompletedUploads chan string
	IgnoredUploads   chan string
	Errors           chan error
}

// Creates a new ConcurrentUploader using the specified credentials.
// The second argument is the id of the album in which images are going to be added when uploaded. Use an empty string
// if you don't want to move the images in to a specific album. The third argument is the maximum number of concurrent
// uploads (which must not be 0).
func NewUploader(credentials auth.CookieCredentials, albumId string, maxConcurrentUploads int) (*ConcurrentUploader, error) {
	if maxConcurrentUploads <= 0 {
		return nil, fmt.Errorf("maxConcurrentUploads must be greater than zero")
	}

	return &ConcurrentUploader{
		credentials: credentials,
		albumId:     albumId,

		concurrentLimiter: make(chan bool, maxConcurrentUploads),

		uploadedFiles: make(map[string]bool),

		CompletedUploads: make(chan string),
		IgnoredUploads:   make(chan string),
		Errors:           make(chan error),
	}, nil
}

// Add files to the list of already uploaded files
func (u *ConcurrentUploader) AddUploadedFiles(files ...string) {
	for _, name := range files {
		u.uploadedFiles[name] = true
	}
}

// Enqueue a new upload. You must not call this method while waiting for some uploads to finish (The method return an
// error if you try to do it).
// Due to the fact that this method is asynchronous, if nil is return it doesn't mean the the upload was completed:
// for that use the Errors and CompletedUploads channels
func (u *ConcurrentUploader) EnqueueUpload(filePath string) error {
	if u.waiting {
		return fmt.Errorf("can't add new uploads while waiting queued uploads to finish")
	}

	// We need to use the absolute path of the file, to avoid multiple uploads of the same file if the tool is executed
	// from different directories
	if !filepath.IsAbs(filePath) {
		if abs, err := filepath.Abs(filePath); err != nil {
			log.Printf("uploader: Can't get the absolute path of file to upload, using relative path. Error: %v\n", err)
		} else {
			filePath = abs
		}
	}

	if u.wasFileAlreadyUploaded(filePath) {
		u.IgnoredUploads <- filePath
		return nil
	}

	// Check if the file is an image or a video
	if valid, err := IsImageOrVideo(filePath); err != nil {
		u.sendError(filePath, err)
		return nil
	} else if !valid {
		u.IgnoredUploads <- filePath
		return nil
	}

	started := make(chan bool)
	go u.uploadFile(filePath, started)
	<-started

	return nil
}

func (u *ConcurrentUploader) wasFileAlreadyUploaded(filePath string) bool {
	_, uploaded := u.uploadedFiles[filePath]
	return uploaded
}

func (u *ConcurrentUploader) uploadFile(filePath string, started chan bool) {
	started <- true
	u.joinGroupAndWaitForTurn()
	defer u.leaveGroupAndNotifyNextUpload()

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		u.sendError(filePath, err)
		return
	}
	defer file.Close()

	// Create options
	options, err := api.NewUploadOptionsFromFile(file)
	if err != nil {
		u.sendError(filePath, err)
		return
	}
	options.AlbumId = u.albumId

	// Create a new upload
	upload, err := api.NewUpload(options, u.credentials)
	if err != nil {
		u.sendError(filePath, err)
		return
	}

	// Try to upload the image
	if _, err := upload.Upload(); err != nil {
		u.sendError(filePath, err)
	} else {
		u.uploadedFiles[filePath] = true
		u.CompletedUploads <- filePath
	}
}

func (u *ConcurrentUploader) sendError(filePath string, err error) {
	u.Errors <- fmt.Errorf("Error with '%s': %s\n", filePath, err)
}

func (u *ConcurrentUploader) joinGroupAndWaitForTurn() {
	u.waitingGroup.Add(1)

	// Insert something in the channel. We remove values from it only when we complete an upload, blocking the
	// goroutines if we exceed the maxConcurrentUpload
	u.concurrentLimiter <- true
}

func (u *ConcurrentUploader) leaveGroupAndNotifyNextUpload() {
	u.waitingGroup.Done()

	// Remove a value to empty the channel or to unlock a waiting gorutine
	<-u.concurrentLimiter
}

// Blocks this goroutine until all the upload are completed. You can not add uploads when a goroutine call this method
func (u *ConcurrentUploader) WaitUploadsCompleted() {
	u.waiting = true
	u.waitingGroup.Wait()
	u.waiting = false
}
